// views/engine.ts — Handlebars template engine initialization and rendering

import Handlebars from "handlebars";
import type { PageData } from "../content/data.ts";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface HandlebarsEngine {
  render(template: string, data: Record<string, unknown>): string;
  registerPartial(name: string, template: string): void;
  registerHelper(name: string, fn: Handlebars.HelperDelegate): void;
  templates: Map<string, Handlebars.TemplateDelegate>;
  partials: Map<string, string>;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Register custom Handlebars helpers.
 */
function registerHelpers(hbs: typeof Handlebars): void {
  // Equality check helper: {{#if (eq a b)}} or {{eq a b}}
  hbs.registerHelper("eq", function (a: unknown, b: unknown): boolean {
    return a === b;
  });

  // Returns the current year for copyright notices
  hbs.registerHelper("currentYear", function (): number {
    return new Date().getFullYear();
  });

  // Maps phase status to a CSS class name
  hbs.registerHelper(
    "statusClass",
    function (status: string): string {
      switch (status) {
        case "complete":
          return "phase-complete";
        case "in-progress":
          return "phase-in-progress";
        case "planned":
          return "phase-planned";
        default:
          return "phase-unknown";
      }
    },
  );
}

// ─── File Loading Utilities ───────────────────────────────────────────────────

/**
 * Load all .hbs files from a directory. Returns a map of filename (without
 * extension) to file content. Gracefully returns an empty map if the directory
 * does not exist or is empty.
 */
async function loadTemplatesFromDir(
  dirPath: string,
): Promise<Map<string, string>> {
  const templates = new Map<string, string>();

  try {
    for await (const entry of Deno.readDir(dirPath)) {
      if (entry.isFile && entry.name.endsWith(".hbs")) {
        const name = entry.name.replace(/\.hbs$/, "");
        const content = await Deno.readTextFile(`${dirPath}/${entry.name}`);
        templates.set(name, content);
      }
    }
  } catch (error) {
    // Directory may not exist yet (templates created in later tasks)
    if (!(error instanceof Deno.errors.NotFound)) {
      throw error;
    }
  }

  return templates;
}

// ─── Engine Initialization ────────────────────────────────────────────────────

/**
 * Initialize the Handlebars engine: load templates, register partials and
 * helpers, and return the configured engine instance.
 */
export async function initHandlebars(
  basePath = "./views",
): Promise<HandlebarsEngine> {
  const hbs = Handlebars.create();

  // Register custom helpers
  registerHelpers(hbs);

  // Load layout templates
  const layouts = await loadTemplatesFromDir(`${basePath}/layouts`);

  // Load and register partials
  const partials = await loadTemplatesFromDir(`${basePath}/partials`);
  for (const [name, content] of partials) {
    hbs.registerPartial(name, content);
  }

  // Register layouts as partials so they can be referenced via {{> main}}
  for (const [name, content] of layouts) {
    hbs.registerPartial(name, content);
  }

  // Load and compile page templates (including subdirectories)
  const pages = await loadTemplatesFromDir(`${basePath}/pages`);
  const compiledTemplates = new Map<string, Handlebars.TemplateDelegate>();
  for (const [name, content] of pages) {
    compiledTemplates.set(name, hbs.compile(content));
  }

  // Load templates from subdirectories (e.g., pages/admin/)
  try {
    for await (const entry of Deno.readDir(`${basePath}/pages`)) {
      if (entry.isDirectory) {
        const subPages = await loadTemplatesFromDir(
          `${basePath}/pages/${entry.name}`,
        );
        for (const [name, content] of subPages) {
          compiledTemplates.set(`${entry.name}/${name}`, hbs.compile(content));
        }
      }
    }
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      throw error;
    }
  }

  // Also compile the main layout for use in renderPage
  const mainLayoutSource = layouts.get("main") ?? "{{{body}}}";
  const mainLayout = hbs.compile(mainLayoutSource);

  const engine: HandlebarsEngine = {
    render(template: string, data: Record<string, unknown>): string {
      const compiled = hbs.compile(template);
      return compiled(data);
    },
    registerPartial(name: string, template: string): void {
      hbs.registerPartial(name, template);
      partials.set(name, template);
    },
    registerHelper(name: string, fn: Handlebars.HelperDelegate): void {
      hbs.registerHelper(name, fn);
    },
    templates: compiledTemplates,
    partials,
  };

  // Store the main layout on the engine for renderPage to use
  (engine as unknown as { _mainLayout: Handlebars.TemplateDelegate })
    ._mainLayout = mainLayout;
  (engine as unknown as { _hbs: typeof Handlebars })._hbs = hbs;

  return engine;
}

// ─── Page Rendering ───────────────────────────────────────────────────────────

/**
 * Generate HTML meta tags string from a PageMeta object.
 */
function generateMetaTags(meta: PageData["meta"]): string {
  const tags: string[] = [];

  if (meta.title) {
    tags.push(`<title>${escapeHtml(meta.title)}</title>`);
  }
  if (meta.description) {
    tags.push(
      `<meta name="description" content="${escapeHtml(meta.description)}">`,
    );
  }
  if (meta.keywords && meta.keywords.length > 0) {
    tags.push(
      `<meta name="keywords" content="${escapeHtml(meta.keywords.join(", "))}">`,
    );
  }
  if (meta.ogTitle) {
    tags.push(
      `<meta property="og:title" content="${escapeHtml(meta.ogTitle)}">`,
    );
  }
  if (meta.ogDescription) {
    tags.push(
      `<meta property="og:description" content="${escapeHtml(meta.ogDescription)}">`,
    );
  }
  if (meta.canonicalUrl) {
    tags.push(
      `<meta property="og:url" content="${escapeHtml(meta.canonicalUrl)}">`,
    );
    tags.push(
      `<link rel="canonical" href="${escapeHtml(meta.canonicalUrl)}">`,
    );
  }
  if (meta.ogImage) {
    tags.push(
      `<meta property="og:image" content="${escapeHtml(meta.ogImage)}">`,
    );
  }

  return tags.join("\n  ");
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

/**
 * Render a page template wrapped in the main layout.
 *
 * The page template is rendered first to produce the body content, then the
 * main layout is rendered with the body content injected as `{{{body}}}`.
 */
export function renderPage(
  engine: HandlebarsEngine,
  page: string,
  data: PageData,
): string {
  const templateFn = engine.templates.get(page);
  if (!templateFn) {
    throw new Error(
      `Template "${page}" not found. Available templates: ${[...engine.templates.keys()].join(", ") || "(none)"}`,
    );
  }

  // Render the page content
  const body = templateFn(data as unknown as Record<string, unknown>);

  // Generate meta tags HTML from the meta object
  const metaTags = generateMetaTags(data.meta);

  // Render the main layout with the page body injected
  const mainLayout = (
    engine as unknown as { _mainLayout: Handlebars.TemplateDelegate }
  )._mainLayout;

  return mainLayout({
    ...data,
    body,
    metaTags,
  } as unknown as Record<string, unknown>);
}
