// tests/seo/seo_test.ts — Property-based tests for SEO module
import { assertEquals, assert } from "https://deno.land/std@0.224.0/assert/mod.ts";
import * as fc from "fast-check";
import { generateSitemap } from "../../content/seo.ts";

/**
 * **Validates: Requirements 9.2**
 *
 * Property 7: Sitemap generation correctness
 *
 * For any non-empty list of page URL paths and a valid base URL,
 * the generateSitemap function SHALL produce valid XML containing
 * a <url><loc> entry for each page path, and the total number of
 * <url> entries SHALL equal the number of input paths.
 */
Deno.test("Property 7: Sitemap generation correctness", () => {
  // Arbitrary for URL-safe path segments
  const pathSegmentArb = fc.stringOf(
    fc.constantFrom(
      ..."abcdefghijklmnopqrstuvwxyz0123456789-_".split("")
    ),
    { minLength: 1, maxLength: 20 }
  );

  // Arbitrary for a URL path like "/foo/bar"
  const urlPathArb = fc
    .array(pathSegmentArb, { minLength: 1, maxLength: 4 })
    .map((segments) => "/" + segments.join("/"));

  // Arbitrary for a valid base URL (scheme + domain)
  const baseUrlArb = fc
    .tuple(
      fc.constantFrom("https://", "http://"),
      fc.stringOf(
        fc.constantFrom(..."abcdefghijklmnopqrstuvwxyz0123456789".split("")),
        { minLength: 3, maxLength: 15 }
      ),
      fc.constantFrom(".com", ".org", ".net", ".io", ".dev")
    )
    .map(([scheme, domain, tld]) => `${scheme}${domain}${tld}`);

  // Arbitrary for a non-empty list of paths
  const pathsArb = fc.array(urlPathArb, { minLength: 1, maxLength: 50 });

  fc.assert(
    fc.property(baseUrlArb, pathsArb, (baseUrl, paths) => {
      const sitemap = generateSitemap(baseUrl, paths);

      // 1. Verify XML declaration is present
      assert(
        sitemap.startsWith('<?xml version="1.0" encoding="UTF-8"?>'),
        "Sitemap must start with XML declaration"
      );

      // 2. Verify urlset root element with correct namespace
      assert(
        sitemap.includes('<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">'),
        "Sitemap must contain urlset element with sitemap namespace"
      );

      // 3. Verify closing urlset tag
      assert(
        sitemap.includes("</urlset>"),
        "Sitemap must contain closing urlset tag"
      );

      // 4. Count <url> entries matches input paths length
      const urlMatches = sitemap.match(/<url>/g);
      const urlCount = urlMatches ? urlMatches.length : 0;
      assertEquals(
        urlCount,
        paths.length,
        `Expected ${paths.length} <url> entries, got ${urlCount}`
      );

      // 5. Count <loc> entries matches input paths length
      const locMatches = sitemap.match(/<loc>/g);
      const locCount = locMatches ? locMatches.length : 0;
      assertEquals(
        locCount,
        paths.length,
        `Expected ${paths.length} <loc> entries, got ${locCount}`
      );

      // 6. Each path should appear as a <loc> entry with the base URL
      const trimmedBase = baseUrl.replace(/\/$/, "");
      for (const path of paths) {
        const normalizedPath = path.startsWith("/") ? path : `/${path}`;
        const expectedUrl = trimmedBase + normalizedPath;
        // The URL is XML-escaped in the output
        const escapedUrl = expectedUrl
          .replace(/&/g, "&amp;")
          .replace(/</g, "&lt;")
          .replace(/>/g, "&gt;")
          .replace(/"/g, "&quot;")
          .replace(/'/g, "&apos;");
        assert(
          sitemap.includes(`<loc>${escapedUrl}</loc>`),
          `Sitemap must contain <loc>${escapedUrl}</loc>`
        );
      }
    }),
    { numRuns: 100 }
  );
});

import { getPageMeta } from "../../content/seo.ts";

// ─── Known pages in the SEO module ───────────────────────────────────────────
const KNOWN_PAGES = [
  "home",
  "roadmap",
  "conops",
  "docs",
  "resources",
  "getting-started",
  "contributing",
  "contact",
  "privacy",
];

/**
 * **Validates: Requirements 9.1, 9.4**
 *
 * Property 6: Page metadata completeness
 *
 * For any valid page name, the returned PageMeta SHALL have non-empty title,
 * description, keywords (at least 1), ogTitle, ogDescription, and canonicalUrl.
 * For any unknown page name, the fallback metadata SHALL also satisfy these
 * completeness requirements.
 */
Deno.test("Property 6: Page metadata completeness — known pages", () => {
  // Arbitrary that picks from the set of known page identifiers
  const knownPageArb = fc.constantFrom(...KNOWN_PAGES);

  fc.assert(
    fc.property(knownPageArb, (page) => {
      const meta = getPageMeta(page);

      // title must be non-empty
      assert(
        meta.title.length > 0,
        `Page "${page}": title must be non-empty`
      );

      // description must be non-empty
      assert(
        meta.description.length > 0,
        `Page "${page}": description must be non-empty`
      );

      // keywords must have at least one entry
      assert(
        meta.keywords.length >= 1,
        `Page "${page}": keywords must have at least 1 entry`
      );

      // All keywords must be non-empty strings
      for (const kw of meta.keywords) {
        assert(
          kw.length > 0,
          `Page "${page}": each keyword must be non-empty`
        );
      }

      // ogTitle must be non-empty
      assert(
        meta.ogTitle.length > 0,
        `Page "${page}": ogTitle must be non-empty`
      );

      // ogDescription must be non-empty
      assert(
        meta.ogDescription.length > 0,
        `Page "${page}": ogDescription must be non-empty`
      );

      // canonicalUrl must be non-empty
      assert(
        meta.canonicalUrl.length > 0,
        `Page "${page}": canonicalUrl must be non-empty`
      );
    }),
    { numRuns: 100 }
  );
});

Deno.test("Property 6: Page metadata completeness — fallback for unknown pages", () => {
  // Arbitrary for random strings that are NOT known page names
  const unknownPageArb = fc
    .stringOf(
      fc.constantFrom(
        ..."abcdefghijklmnopqrstuvwxyz0123456789-_".split("")
      ),
      { minLength: 1, maxLength: 30 }
    )
    .filter((s) => !KNOWN_PAGES.includes(s));

  fc.assert(
    fc.property(unknownPageArb, (page) => {
      const meta = getPageMeta(page);

      // Fallback metadata must still be complete

      // title must be non-empty
      assert(
        meta.title.length > 0,
        `Fallback for "${page}": title must be non-empty`
      );

      // description must be non-empty
      assert(
        meta.description.length > 0,
        `Fallback for "${page}": description must be non-empty`
      );

      // keywords must have at least one entry
      assert(
        meta.keywords.length >= 1,
        `Fallback for "${page}": keywords must have at least 1 entry`
      );

      // All keywords must be non-empty strings
      for (const kw of meta.keywords) {
        assert(
          kw.length > 0,
          `Fallback for "${page}": each keyword must be non-empty`
        );
      }

      // ogTitle must be non-empty
      assert(
        meta.ogTitle.length > 0,
        `Fallback for "${page}": ogTitle must be non-empty`
      );

      // ogDescription must be non-empty
      assert(
        meta.ogDescription.length > 0,
        `Fallback for "${page}": ogDescription must be non-empty`
      );

      // canonicalUrl must be non-empty
      assert(
        meta.canonicalUrl.length > 0,
        `Fallback for "${page}": canonicalUrl must be non-empty`
      );
    }),
    { numRuns: 100 }
  );
});
