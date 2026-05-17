// routes/admin-statistics.ts — Admin web access statistics route handler

import type { RouterContext } from "@oak/oak";
import type { HandlebarsEngine } from "../views/engine.ts";
import { renderPage } from "../views/engine.ts";
import { siteContent } from "../content/data.ts";
import { getPageMeta } from "../content/seo.ts";
import type { PageData } from "../content/data.ts";
import { getTemplateUser } from "./helpers.ts";
import {
  parseLogFiles,
  generateStatistics,
  formatBytes,
} from "../services/log-parser.ts";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function buildAdminPageData(
  activeSection: string,
  content: Record<string, unknown>,
  options?: {
    errors?: string[];
    user?: {
      id: number;
      name: string;
      email: string;
      roles: Array<{ id: number; name: string; description: string }>;
      isAdmin?: boolean;
    } | null;
  },
): PageData & { errors?: string[] } {
  const meta = getPageMeta("admin");
  const nav = siteContent.nav.map((item) => ({
    ...item,
    active: false,
  }));

  return {
    meta,
    nav,
    activeSection,
    content,
    collaborators: siteContent.overview.collaborators,
    currentYear: new Date().getFullYear(),
    user: options?.user ?? null,
    ...(options?.errors ? { errors: options.errors } : {}),
  };
}

/**
 * Get a CSS class for an HTTP status code badge.
 */
function getStatusBadgeClass(status: number): string {
  if (status >= 200 && status < 300) return "bg-success";
  if (status >= 300 && status < 400) return "bg-info";
  if (status >= 400 && status < 500) return "bg-warning text-dark";
  if (status >= 500) return "bg-danger";
  return "bg-secondary";
}

// ─── Handler Factory ──────────────────────────────────────────────────────────

/**
 * GET /admin/statistics — Render web access statistics page.
 * Requires requireAdmin() middleware applied externally.
 */
export function statisticsHandler(engine: HandlebarsEngine) {
  return async (ctx: RouterContext<"/admin/statistics">) => {
    const url = ctx.request.url;
    const includeOwner = url.searchParams.get("include_owner") === "true";

    try {
      // Parse log files
      const entries = await parseLogFiles();

      // Get the requesting user's IP as the "owner" IP to exclude
      // Use X-Forwarded-For if behind a proxy, otherwise use the direct IP
      const ownerIP = ctx.request.ip;

      // Generate statistics
      const stats = generateStatistics(entries, !includeOwner, ownerIP);

      // Format data for template
      const totalRequests = stats.totalRequests;
      const humanPercent = totalRequests > 0
        ? Math.round((stats.humanRequests / totalRequests) * 100)
        : 0;
      const botPercent = totalRequests > 0
        ? Math.round((stats.botRequests / totalRequests) * 100)
        : 0;

      // Convert requestsByHour map to array for template
      const maxHourRequests = Math.max(...[...stats.requestsByHour.values()], 1);
      const hourlyData = Array.from({ length: 24 }, (_, hour) => {
        const count = stats.requestsByHour.get(hour) ?? 0;
        const heightPercent = Math.round((count / maxHourRequests) * 100);
        return {
          hour: hour.toString().padStart(2, "0"),
          count,
          heightPercent: Math.max(heightPercent, 2), // minimum bar height for visibility
        };
      });

      // Convert requestsByDay map to sorted array (last 30 days)
      const sortedDays = [...stats.requestsByDay.entries()]
        .sort((a, b) => a[0].localeCompare(b[0]))
        .slice(-30);
      const maxDayRequests = Math.max(
        ...sortedDays.map(([, count]) => count),
        1,
      );
      const dailyData = sortedDays.map(([day, count]) => ({
        day,
        count,
        widthPercent: Math.round((count / maxDayRequests) * 100),
      }));

      // Convert statusCodes map to array with badge classes
      const statusCodesData = [...stats.statusCodes.entries()]
        .sort((a, b) => a[0] - b[0])
        .map(([code, count]) => ({
          code,
          count,
          badgeClass: getStatusBadgeClass(code),
          percent: totalRequests > 0
            ? Math.round((count / totalRequests) * 100)
            : 0,
        }));

      // Convert browsers map to sorted array
      const totalHuman = stats.humanRequests || 1;
      const browsersData = [...stats.browsers.entries()]
        .sort((a, b) => b[1] - a[1])
        .map(([name, count]) => ({
          name,
          count,
          percent: Math.round((count / totalHuman) * 100),
        }));

      const content = {
        totalRequests: stats.totalRequests.toLocaleString(),
        uniqueVisitors: stats.uniqueVisitors.toLocaleString(),
        pageViews: stats.pageViews.toLocaleString(),
        bandwidth: formatBytes(stats.bandwidth),
        humanRequests: stats.humanRequests.toLocaleString(),
        botRequests: stats.botRequests.toLocaleString(),
        humanPercent,
        botPercent,
        topPages: stats.topPages,
        topIPs: stats.topIPs,
        statusCodes: statusCodesData,
        browsers: browsersData,
        hourlyData,
        dailyData,
        includeOwner,
      };

      const pageData = buildAdminPageData("admin-statistics", content, {
        user: getTemplateUser(ctx),
      });

      const html = renderPage(engine, "admin/statistics", pageData);
      ctx.response.body = html;
      ctx.response.type = "text/html";
    } catch (error) {
      console.error("Error generating statistics:", error);

      const content = {
        error: error instanceof Error
          ? error.message
          : "An unexpected error occurred while reading access logs.",
        totalRequests: "0",
        uniqueVisitors: "0",
        pageViews: "0",
        bandwidth: "0 B",
        humanRequests: "0",
        botRequests: "0",
        humanPercent: 0,
        botPercent: 0,
        topPages: [],
        topIPs: [],
        statusCodes: [],
        browsers: [],
        hourlyData: [],
        dailyData: [],
        includeOwner,
      };

      const pageData = buildAdminPageData("admin-statistics", content, {
        errors: ["Failed to load access log statistics. Check server logs for details."],
        user: getTemplateUser(ctx),
      });

      const html = renderPage(engine, "admin/statistics", pageData);
      ctx.response.body = html;
      ctx.response.type = "text/html";
    }
  };
}
