// content/seo.ts — SEO metadata and sitemap generation

// ─── Interfaces ───────────────────────────────────────────────────────────────

export interface PageMeta {
  title: string;
  description: string;
  keywords: string[];
  ogTitle: string;
  ogDescription: string;
  ogImage?: string;
  canonicalUrl: string;
}

// ─── Per-Page Metadata ────────────────────────────────────────────────────────

const BASE_KEYWORDS = [
  "DTN",
  "Delay-Tolerant Networking",
  "amateur radio",
  "Bundle Protocol",
  "cislunar communication",
  "BPv7",
  "HDTN",
  "space networking",
];

const pageMeta: Record<string, PageMeta> = {
  home: {
    title: "Amateur Radio DTN Space Networking Pathfinder",
    description:
      "From amateur packet radio to CubeSat relay to cislunar networking. " +
      "Bringing Delay-Tolerant Networking and Bundle Protocol to amateur radio links.",
    keywords: [
      ...BASE_KEYWORDS,
      "AMSAT",
      "CubeSat",
      "store-and-forward",
      "AX.25",
    ],
    ogTitle: "Amateur Radio DTN Space Networking Pathfinder",
    ogDescription:
      "Open-source project bringing DTN and Bundle Protocol to amateur radio, " +
      "from terrestrial links to cislunar space.",
    canonicalUrl: "/",
  },
  roadmap: {
    title: "Project Roadmap — Amateur Radio DTN Pathfinder",
    description:
      "Five-phase roadmap from terrestrial DTN validation through QO-100 GEO satellite, " +
      "CubeSat engineering model, LEO flight, to cislunar deep-space communication.",
    keywords: [
      ...BASE_KEYWORDS,
      "roadmap",
      "CubeSat",
      "LEO",
      "QO-100",
      "deep-space",
    ],
    ogTitle: "Project Roadmap — Amateur Radio DTN Pathfinder",
    ogDescription:
      "From ground-based amateur radio DTN to cislunar deep-space networking in five phases.",
    canonicalUrl: "/roadmap",
  },
  conops: {
    title: "Concept of Operations — Amateur Radio DTN Pathfinder",
    description:
      "Concept of operations for the terrestrial analogue architecture mirroring " +
      "cislunar communications using amateur radio DTN links.",
    keywords: [
      ...BASE_KEYWORDS,
      "ConOps",
      "concept of operations",
      "terrestrial analogue",
      "RF links",
    ],
    ogTitle: "Concept of Operations — Amateur Radio DTN Pathfinder",
    ogDescription:
      "Terrestrial analogue architecture for cislunar DTN operations using amateur radio.",
    canonicalUrl: "/conops",
  },
  docs: {
    title: "Documentation — Amateur Radio DTN Pathfinder",
    description:
      "Technical documentation including phase-specific requirements, design documents, " +
      "RFC references for Bundle Protocol (BPv7), LTP, and AX.25.",
    keywords: [
      ...BASE_KEYWORDS,
      "documentation",
      "RFC 9171",
      "RFC 5326",
      "AX.25",
      "LTP",
    ],
    ogTitle: "Documentation — Amateur Radio DTN Pathfinder",
    ogDescription:
      "Phase-specific documentation and protocol references for the amateur radio DTN project.",
    canonicalUrl: "/docs",
  },
  resources: {
    title: "Resources — Amateur Radio DTN Pathfinder",
    description:
      "External resources including NASA DTN references, HDTN software, " +
      "Bundle Protocol RFCs, AMSAT organisations, and AX.25/KISS protocol information.",
    keywords: [
      ...BASE_KEYWORDS,
      "resources",
      "NASA",
      "AMSAT-UK",
      "AMSAT-DL",
      "KISS",
    ],
    ogTitle: "Resources — Amateur Radio DTN Pathfinder",
    ogDescription:
      "Curated links to NASA DTN, HDTN, Bundle Protocol, AMSAT, and AX.25 resources.",
    canonicalUrl: "/resources",
  },
  "getting-started": {
    title: "Getting Started — Amateur Radio DTN Pathfinder",
    description:
      "Step-by-step guide to setting up a terrestrial DTN node using amateur radio " +
      "equipment: Raspberry Pi, Mobilinkd TNC4, and Yaesu FT-817 with Bundle Protocol.",
    keywords: [
      ...BASE_KEYWORDS,
      "getting started",
      "installation",
      "TNC4",
      "Yaesu FT-817",
      "Raspberry Pi",
    ],
    ogTitle: "Getting Started — Amateur Radio DTN Pathfinder",
    ogDescription:
      "Set up your own terrestrial DTN node with amateur radio equipment and Bundle Protocol.",
    canonicalUrl: "/getting-started",
  },
  contributing: {
    title: "Contributing — Amateur Radio DTN Pathfinder",
    description:
      "How to contribute to the amateur radio DTN project: bug reports, documentation, " +
      "hardware integration, RF optimisation, and orbital mechanics.",
    keywords: [
      ...BASE_KEYWORDS,
      "contributing",
      "open source",
      "MIT license",
    ],
    ogTitle: "Contributing — Amateur Radio DTN Pathfinder",
    ogDescription:
      "Join the project: contribute code, hardware expertise, RF knowledge, or documentation.",
    canonicalUrl: "/contributing",
  },
  contact: {
    title: "Contact — Amateur Radio DTN Pathfinder",
    description:
      "Get in touch with the Cislunar Amateur DTN project team. Collaboration enquiries, " +
      "amateur radio contact, AMSAT-UK, AMSAT-DL partnerships.",
    keywords: [
      ...BASE_KEYWORDS,
      "contact",
      "collaboration",
      "AMSAT-UK",
      "AMSAT-DL",
      "partnership",
    ],
    ogTitle: "Contact — Amateur Radio DTN Pathfinder",
    ogDescription:
      "Reach the project team for collaboration, partnership, or general enquiries.",
    canonicalUrl: "/contact",
  },
  privacy: {
    title: "Privacy Policy — Amateur Radio DTN Pathfinder",
    description:
      "GDPR privacy statement for the Amateur Radio DTN Pathfinder website. " +
      "Data collection, legal basis, retention, and cookie policy.",
    keywords: ["privacy", "GDPR", "data protection", "cookies"],
    ogTitle: "Privacy Policy — Amateur Radio DTN Pathfinder",
    ogDescription:
      "How we handle your data: GDPR-compliant privacy statement.",
    canonicalUrl: "/privacy",
  },
};

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Returns SEO metadata for a given page identifier.
 * Falls back to homepage meta if the page is not recognised.
 */
export function getPageMeta(page: string): PageMeta {
  const normalised = page.replace(/^\//, "").replace(/\/$/, "") || "home";
  return Object.hasOwn(pageMeta, normalised)
    ? pageMeta[normalised]!
    : pageMeta["home"]!;
}

/**
 * Generates a valid XML sitemap string for the given base URL and page paths.
 */
export function generateSitemap(baseUrl: string, pages: string[]): string {
  const trimmedBase = baseUrl.replace(/\/$/, "");

  const urls = pages
    .map((page) => {
      const path = page.startsWith("/") ? page : `/${page}`;
      return (
        `  <url>\n` +
        `    <loc>${escapeXml(trimmedBase + path)}</loc>\n` +
        `  </url>`
      );
    })
    .join("\n");

  return (
    `<?xml version="1.0" encoding="UTF-8"?>\n` +
    `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n` +
    urls +
    `\n</urlset>\n`
  );
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function escapeXml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}
