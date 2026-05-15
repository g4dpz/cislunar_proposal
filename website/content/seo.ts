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
    title: "RADIANT — Radio Amateur Delay-tolerant Interplanetary Networking Testbed",
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
    ogTitle: "RADIANT — Radio Amateur Delay-tolerant Interplanetary Networking Testbed",
    ogDescription:
      "Open-source project bringing DTN and Bundle Protocol to amateur radio, " +
      "from terrestrial links to cislunar space.",
    canonicalUrl: "/",
  },
  roadmap: {
    title: "Project Roadmap — RADIANT",
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
    ogTitle: "Project Roadmap — RADIANT",
    ogDescription:
      "From ground-based amateur radio DTN to cislunar deep-space networking in five phases.",
    canonicalUrl: "/roadmap",
  },
  conops: {
    title: "Concept of Operations — RADIANT",
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
    ogTitle: "Concept of Operations — RADIANT",
    ogDescription:
      "Terrestrial analogue architecture for cislunar DTN operations using amateur radio.",
    canonicalUrl: "/conops",
  },
  docs: {
    title: "Documentation — RADIANT",
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
    ogTitle: "Documentation — RADIANT",
    ogDescription:
      "Phase-specific documentation and protocol references for the amateur radio DTN project.",
    canonicalUrl: "/docs",
  },
  contact: {
    title: "Contact — RADIANT",
    description:
      "Get in touch with the RADIANT project team. Collaboration enquiries, " +
      "amateur radio contact, AMSAT-UK, AMSAT-DL partnerships.",
    keywords: [
      ...BASE_KEYWORDS,
      "contact",
      "collaboration",
      "AMSAT-UK",
      "AMSAT-DL",
      "partnership",
    ],
    ogTitle: "Contact — RADIANT",
    ogDescription:
      "Reach the project team for collaboration, partnership, or general enquiries.",
    canonicalUrl: "/contact",
  },
  privacy: {
    title: "Privacy Policy — RADIANT",
    description:
      "GDPR privacy statement for the RADIANT website. " +
      "Data collection, legal basis, retention, and cookie policy.",
    keywords: ["privacy", "GDPR", "data protection", "cookies"],
    ogTitle: "Privacy Policy — RADIANT",
    ogDescription:
      "How we handle your data: GDPR-compliant privacy statement.",
    canonicalUrl: "/privacy",
  },
  profile: {
    title: "My Profile — RADIANT",
    description:
      "View and manage your RADIANT account profile, update your name and email, " +
      "or change your password.",
    keywords: [...BASE_KEYWORDS, "profile", "account", "settings"],
    ogTitle: "My Profile — RADIANT",
    ogDescription:
      "Manage your RADIANT account settings and profile information.",
    canonicalUrl: "/profile",
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
