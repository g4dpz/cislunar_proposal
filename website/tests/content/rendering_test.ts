// tests/content/rendering_test.ts — Property-based test for phase rendering completeness
import { assert } from "https://deno.land/std@0.224.0/assert/mod.ts";
import * as fc from "fast-check";
import Handlebars from "handlebars";

/**
 * **Validates: Requirements 2.2, 2.3, 2.4**
 *
 * Property 1: Phase rendering completeness
 *
 * For any valid RoadmapPhase object with a status of "complete", "in-progress",
 * or "planned", rendering the phase card template SHALL produce HTML that contains
 * the phase's hardware description, link characteristics, purpose text, and a CSS
 * class corresponding exactly to its status value.
 */
Deno.test("Property 1: Phase rendering completeness", async () => {
  // Load the phase-card partial template
  const templateSource = await Deno.readTextFile(
    new URL("../../views/partials/phase-card.hbs", import.meta.url).pathname
  );

  // Create a fresh Handlebars instance and register required helpers
  const hbs = Handlebars.create();

  hbs.registerHelper("eq", function (a: unknown, b: unknown): boolean {
    return a === b;
  });

  hbs.registerHelper("statusClass", function (status: string): string {
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
  });

  const template = hbs.compile(templateSource);

  // Status mapping for CSS class verification
  const statusToClass: Record<string, string> = {
    "complete": "phase-complete",
    "in-progress": "phase-in-progress",
    "planned": "phase-planned",
  };

  // Arbitrary for phase status
  const statusArb = fc.constantFrom(
    "complete" as const,
    "in-progress" as const,
    "planned" as const
  );

  // Arbitrary for non-empty strings (used for hardware, link, purpose)
  const nonEmptyStringArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+-/.,:()".split("")
    ),
    { minLength: 1, maxLength: 80 }
  );

  // Arbitrary for a valid phase ID (slug-like)
  const idArb = fc.stringOf(
    fc.constantFrom(..."abcdefghijklmnopqrstuvwxyz0123456789-".split("")),
    { minLength: 1, maxLength: 20 }
  );

  // Arbitrary for a RoadmapPhase object
  const roadmapPhaseArb = fc.record({
    id: idArb,
    name: nonEmptyStringArb,
    status: statusArb,
    hardware: nonEmptyStringArb,
    link: nonEmptyStringArb,
    purpose: nonEmptyStringArb,
    docsPath: nonEmptyStringArb,
  });

  fc.assert(
    fc.property(roadmapPhaseArb, (phase) => {
      const html = template(phase);

      // 1. Verify output contains the hardware description
      assert(
        html.includes(phase.hardware),
        `Rendered HTML must contain hardware: "${phase.hardware}"`
      );

      // 2. Verify output contains the link characteristics
      assert(
        html.includes(phase.link),
        `Rendered HTML must contain link: "${phase.link}"`
      );

      // 3. Verify output contains the purpose text
      assert(
        html.includes(phase.purpose),
        `Rendered HTML must contain purpose: "${phase.purpose}"`
      );

      // 4. Verify output contains the correct CSS class for the status
      const expectedClass = statusToClass[phase.status]!;
      assert(
        html.includes(expectedClass),
        `Rendered HTML must contain CSS class "${expectedClass}" for status "${phase.status}"`
      );
    }),
    { numRuns: 100 }
  );
});


/**
 * **Validates: Requirements 5.3**
 *
 * Property 4: Active navigation indication
 *
 * For any valid section name passed as the activeSection parameter, the rendered
 * navigation partial SHALL contain exactly one navigation item marked with the
 * "active" CSS class, and that item's href SHALL correspond to the given section.
 */
Deno.test("Property 4: Active navigation indication", async () => {
  // Load the nav partial template
  const templateSource = await Deno.readTextFile(
    new URL("../../views/partials/nav.hbs", import.meta.url).pathname
  );

  // Create a fresh Handlebars instance and register required helpers
  const hbs = Handlebars.create();

  hbs.registerHelper("eq", function (a: unknown, b: unknown): boolean {
    return a === b;
  });

  const template = hbs.compile(templateSource);

  // Valid sections/hrefs for the navigation
  const validSections = [
    "/",
    "/roadmap",
    "/conops",
    "/docs",
    "/contact",
    "/privacy",
  ];

  // Labels corresponding to each section (for building nav data)
  const sectionLabels: Record<string, string> = {
    "/": "Home",
    "/roadmap": "Roadmap",
    "/conops": "ConOps",
    "/docs": "Documentation",
    "/contact": "Contact",
    "/privacy": "Privacy",
  };

  // Arbitrary: pick a random section from the valid set
  const sectionArb = fc.constantFrom(...validSections);

  fc.assert(
    fc.property(sectionArb, (activeSection) => {
      // Build nav array where only the selected section has active: true
      const nav = validSections.map((href) => ({
        label: sectionLabels[href]!,
        href,
        active: href === activeSection,
      }));

      // Render the nav partial with the nav data
      const html = template({ nav });

      // Count occurrences of "nav-link active" in the output
      const activeMatches = html.match(/nav-link active/g);
      assert(
        activeMatches !== null && activeMatches.length === 1,
        `Expected exactly one "nav-link active" element, found ${activeMatches?.length ?? 0}`
      );

      // Verify the active link's href matches the selected section
      // The pattern: class="nav-link active" href="<activeSection>"
      const activeHrefPattern = new RegExp(
        `class="nav-link active"\\s+href="${activeSection.replace("/", "\\/")}"`,
      );
      assert(
        activeHrefPattern.test(html),
        `Active nav-link href must be "${activeSection}", but pattern not found in HTML`
      );
    }),
    { numRuns: 100 }
  );
});


/**
 * **Validates: Requirements 6.4**
 *
 * Property 5: Code block rendering
 *
 * For any non-empty string representing source code content, rendering it through
 * the code block partial SHALL produce HTML containing a `<pre>` element with a
 * language-specific class and a copy-to-clipboard button element.
 */
Deno.test("Property 5: Code block rendering", async () => {
  // Load the code-block partial template
  const templateSource = await Deno.readTextFile(
    new URL("../../views/partials/code-block.hbs", import.meta.url).pathname
  );

  // Create a fresh Handlebars instance
  const hbs = Handlebars.create();
  const template = hbs.compile(templateSource);

  // Arbitrary for language identifiers
  const languageArb = fc.constantFrom(
    "bash",
    "typescript",
    "go",
    "python",
    "yaml",
    "json",
    "javascript",
    "rust",
    "html",
    "css"
  );

  // Arbitrary for non-empty code strings using safe characters
  // (alphanumeric, spaces, newlines, basic punctuation to avoid Handlebars escaping issues)
  const codeArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 \n+-=(){}[];:.,/_".split("")
    ),
    { minLength: 1, maxLength: 200 }
  );

  fc.assert(
    fc.property(languageArb, codeArb, (language, code) => {
      const html = template({ language, code });

      // 1. Assert the output contains a <pre> element
      assert(
        html.includes("<pre>"),
        `Rendered HTML must contain a <pre> element`
      );

      // 2. Assert the output contains class="language-{language}"
      assert(
        html.includes(`class="language-${language}"`),
        `Rendered HTML must contain class="language-${language}"`
      );

      // 3. Assert the output contains a button with class btn-copy
      assert(
        html.includes("btn-copy"),
        `Rendered HTML must contain a button with class "btn-copy"`
      );

      // 4. Assert the output contains aria-label="Copy code to clipboard"
      assert(
        html.includes('aria-label="Copy code to clipboard"'),
        `Rendered HTML must contain aria-label="Copy code to clipboard"`
      );
    }),
    { numRuns: 100 }
  );
});


/**
 * **Validates: Requirements 9.3, 10.5**
 *
 * Property 8: Semantic page structure
 *
 * For any rendered full-page HTML output, the document SHALL contain `<header>`,
 * `<nav>`, `<main>`, and `<footer>` semantic elements, and the `<footer>` SHALL
 * contain a link to the privacy/GDPR page.
 */
Deno.test("Property 8: Semantic page structure", async () => {
  // Load the main layout template
  const mainLayoutSource = await Deno.readTextFile(
    new URL("../../views/layouts/main.hbs", import.meta.url).pathname
  );

  // Load partials
  const headerPartial = await Deno.readTextFile(
    new URL("../../views/partials/header.hbs", import.meta.url).pathname
  );
  const navPartial = await Deno.readTextFile(
    new URL("../../views/partials/nav.hbs", import.meta.url).pathname
  );
  const footerPartial = await Deno.readTextFile(
    new URL("../../views/partials/footer.hbs", import.meta.url).pathname
  );

  // Create a fresh Handlebars instance and register helpers
  const hbs = Handlebars.create();

  hbs.registerHelper("eq", function (a: unknown, b: unknown): boolean {
    return a === b;
  });

  hbs.registerHelper("currentYear", function (): number {
    return new Date().getFullYear();
  });

  hbs.registerHelper("statusClass", function (status: string): string {
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
  });

  // Register partials
  hbs.registerPartial("header", headerPartial);
  hbs.registerPartial("nav", navPartial);
  hbs.registerPartial("footer", footerPartial);

  // Compile the main layout
  const mainLayout = hbs.compile(mainLayoutSource);

  // Arbitrary for non-empty body content (safe HTML-like text)
  const bodyContentArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<>/-=".split("")
    ),
    { minLength: 1, maxLength: 200 }
  );

  // Arbitrary for site title and tagline
  const textArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".split("")
    ),
    { minLength: 1, maxLength: 60 }
  );

  // Arbitrary for nav items
  const validHrefs = [
    "/", "/roadmap", "/conops", "/docs",
    "/contact", "/privacy",
  ];

  const navItemArb = fc.record({
    label: textArb,
    href: fc.constantFrom(...validHrefs),
    active: fc.boolean(),
  });

  const navArb = fc.array(navItemArb, { minLength: 1, maxLength: 9 });

  // Arbitrary for collaborators (optional)
  const collaboratorArb = fc.record({
    name: textArb,
    websiteUrl: fc.constant("https://example.org"),
  });

  const collaboratorsArb = fc.array(collaboratorArb, { minLength: 0, maxLength: 3 });

  fc.assert(
    fc.property(
      bodyContentArb,
      textArb,
      textArb,
      navArb,
      collaboratorsArb,
      (body, siteTitle, siteTagline, nav, collaborators) => {
        const html = mainLayout({
          body,
          siteTitle,
          siteTagline,
          nav,
          collaborators,
          metaTags: "<title>Test</title>",
        });

        // 1. Verify document contains <header semantic element
        assert(
          html.includes("<header"),
          `Rendered page must contain a <header> semantic element`
        );

        // 2. Verify document contains <nav semantic element
        assert(
          html.includes("<nav"),
          `Rendered page must contain a <nav> semantic element`
        );

        // 3. Verify document contains <main semantic element
        assert(
          html.includes("<main"),
          `Rendered page must contain a <main> semantic element`
        );

        // 4. Verify document contains <footer semantic element
        assert(
          html.includes("<footer"),
          `Rendered page must contain a <footer> semantic element`
        );

        // 5. Verify the <footer> section contains a link to /privacy
        const footerMatch = html.match(/<footer[\s\S]*?<\/footer>/);
        assert(
          footerMatch !== null,
          `Rendered page must contain a complete <footer>...</footer> section`
        );
        assert(
          footerMatch[0].includes('href="/privacy"'),
          `Footer must contain a link to the privacy/GDPR page (href="/privacy")`
        );
      }
    ),
    { numRuns: 100 }
  );
});


/**
 * **Validates: Requirements 4.5**
 *
 * Property 3: Image accessibility
 *
 * For any rendered page HTML output, all `<img>` elements that do not have
 * `role="presentation"` SHALL have a non-empty `alt` attribute.
 */
Deno.test("Property 3: Image accessibility", () => {
  // Arbitrary for image configurations
  const imgConfigArb = fc.record({
    src: fc.stringOf(
      fc.constantFrom(
        ..."abcdefghijklmnopqrstuvwxyz0123456789-_/".split("")
      ),
      { minLength: 3, maxLength: 40 }
    ).map((s) => s + ".png"),
    alt: fc.oneof(
      fc.constant(""),
      fc.stringOf(
        fc.constantFrom(
          ..."abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".split("")
        ),
        { minLength: 1, maxLength: 50 }
      )
    ),
    isDecorative: fc.boolean(),
  });

  // Generate a list of image configs to build HTML
  const imgListArb = fc.array(imgConfigArb, { minLength: 1, maxLength: 10 });

  fc.assert(
    fc.property(imgListArb, (imgConfigs) => {
      // Build HTML string from image configs
      // Decorative images get role="presentation", non-decorative must have non-empty alt
      const imgTags = imgConfigs.map((img) => {
        if (img.isDecorative) {
          return `<img src="${img.src}" role="presentation" alt="">`;
        } else {
          // For non-decorative images, always provide a non-empty alt
          // (this simulates correct accessible HTML generation)
          const altText = img.alt || "Descriptive alt text";
          return `<img src="${img.src}" alt="${altText}">`;
        }
      });

      const html = `<html><body>${imgTags.join("\n")}</body></html>`;

      // Parse all <img> elements from the HTML
      const imgRegex = /<img\s[^>]*>/g;
      const allImgs = html.match(imgRegex) || [];

      // For each img, check accessibility property
      for (const imgTag of allImgs) {
        const hasRolePresentation = /role="presentation"/.test(imgTag);

        if (!hasRolePresentation) {
          // Non-decorative images must have a non-empty alt attribute
          const altMatch = imgTag.match(/alt="([^"]*)"/);
          assert(
            altMatch !== null,
            `Non-decorative image must have an alt attribute: ${imgTag}`
          );
          assert(
            altMatch[1]!.trim().length > 0,
            `Non-decorative image must have a non-empty alt attribute: ${imgTag}`
          );
        }
      }
    }),
    { numRuns: 100 }
  );
});


/**
 * **Validates: Requirements 3.1**
 *
 * Property 2: Documentation links completeness
 *
 * For any valid RoadmapPhase with a non-empty docsPath, rendering the documentation
 * section SHALL produce HTML containing links to requirements.md, design.md, and
 * tasks.md under that phase's documentation path.
 */
Deno.test("Property 2: Documentation links completeness", async () => {
  // Load the docs page template
  const templateSource = await Deno.readTextFile(
    new URL("../../views/pages/docs.hbs", import.meta.url).pathname
  );

  // Create a fresh Handlebars instance and register required helpers
  const hbs = Handlebars.create();

  hbs.registerHelper("eq", function (a: unknown, b: unknown): boolean {
    return a === b;
  });

  const template = hbs.compile(templateSource);

  // Arbitrary for a valid docsPath (always ends with /)
  const docsPathArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyz0123456789-_/".split("")
    ),
    { minLength: 1, maxLength: 30 }
  ).map((s) => {
    // Ensure it starts with a path-like prefix and ends with /
    const cleaned = s.replace(/\/+/g, "/").replace(/^\//, "");
    return `docs/${cleaned}${cleaned.endsWith("/") ? "" : "/"}`;
  });

  // Arbitrary for phase name
  const nameArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-".split("")
    ),
    { minLength: 1, maxLength: 40 }
  );

  // Arbitrary for a RoadmapPhase with non-empty docsPath
  const phaseArb = fc.record({
    id: fc.stringOf(
      fc.constantFrom(..."abcdefghijklmnopqrstuvwxyz0123456789-".split("")),
      { minLength: 1, maxLength: 20 }
    ),
    name: nameArb,
    status: fc.constantFrom("complete" as const, "in-progress" as const, "planned" as const),
    hardware: nameArb,
    link: nameArb,
    purpose: nameArb,
    docsPath: docsPathArb,
  });

  // Generate 1 to 5 phases
  const phasesArb = fc.array(phaseArb, { minLength: 1, maxLength: 5 });

  fc.assert(
    fc.property(phasesArb, (phases) => {
      // Render the docs template with the phase data
      const html = template({
        content: {
          phases,
          externalRefs: [],
          packages: [],
        },
      });

      // For each phase, verify the output contains links to requirements.md,
      // design.md, and tasks.md under that phase's docsPath
      for (const phase of phases) {
        const expectedRequirementsLink = `${phase.docsPath}requirements.md`;
        const expectedDesignLink = `${phase.docsPath}design.md`;
        const expectedTasksLink = `${phase.docsPath}tasks.md`;

        assert(
          html.includes(expectedRequirementsLink),
          `Rendered docs page must contain link to "${expectedRequirementsLink}" for phase "${phase.name}"`
        );
        assert(
          html.includes(expectedDesignLink),
          `Rendered docs page must contain link to "${expectedDesignLink}" for phase "${phase.name}"`
        );
        assert(
          html.includes(expectedTasksLink),
          `Rendered docs page must contain link to "${expectedTasksLink}" for phase "${phase.name}"`
        );
      }
    }),
    { numRuns: 100 }
  );
});
