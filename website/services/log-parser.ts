// services/log-parser.ts — Apache access log parser and statistics generator

// ─── Types ────────────────────────────────────────────────────────────────────

export interface LogEntry {
  ip: string;
  timestamp: Date;
  method: string;
  path: string;
  protocol: string;
  status: number;
  bytes: number;
  referer: string;
  userAgent: string;
}

export interface Statistics {
  totalRequests: number;
  uniqueVisitors: number;
  pageViews: number;
  bandwidth: number;
  topPages: Array<{ path: string; count: number }>;
  topIPs: Array<{ ip: string; count: number; isBot: boolean }>;
  statusCodes: Map<number, number>;
  requestsByHour: Map<number, number>;
  requestsByDay: Map<string, number>;
  browsers: Map<string, number>;
  botRequests: number;
  humanRequests: number;
}

// ─── Constants ────────────────────────────────────────────────────────────────

const LOG_PATH = "/var/log/apache2/radiant-access.log";

// Apache Combined Log Format regex
// Example: 192.168.1.1 - - [10/Oct/2023:13:55:36 +0000] "GET /path HTTP/1.1" 200 2326 "http://ref.com" "Mozilla/5.0..."
const LOG_LINE_REGEX =
  /^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)"/;

const BOT_PATTERNS = [
  /bot/i,
  /crawl/i,
  /spider/i,
  /slurp/i,
  /mediapartners/i,
  /facebookexternalhit/i,
  /bingpreview/i,
  /yandex/i,
  /baidu/i,
  /duckduckbot/i,
  /semrush/i,
  /ahrefs/i,
  /mj12bot/i,
  /dotbot/i,
  /petalbot/i,
  /bytespider/i,
  /gptbot/i,
  /claudebot/i,
  /applebot/i,
  /archive\.org/i,
  /wget/i,
  /curl/i,
  /python-requests/i,
  /go-http-client/i,
  /java\//i,
  /libwww/i,
  /scrapy/i,
  /headlesschrome/i,
  /phantomjs/i,
];

const STATIC_EXTENSIONS = [
  ".css",
  ".js",
  ".png",
  ".jpg",
  ".jpeg",
  ".gif",
  ".svg",
  ".ico",
  ".woff",
  ".woff2",
  ".ttf",
  ".eot",
  ".map",
  ".webp",
  ".avif",
];

// ─── Utility Functions ────────────────────────────────────────────────────────

/**
 * Determine if a user agent string belongs to a bot/crawler.
 */
export function isBot(userAgent: string): boolean {
  if (!userAgent || userAgent === "-") return true;
  return BOT_PATTERNS.some((pattern) => pattern.test(userAgent));
}

/**
 * Determine if a request path is a page view (not a static asset or API call).
 */
export function isPageView(path: string): boolean {
  if (isStaticAsset(path)) return false;
  if (path.startsWith("/api/")) return false;
  if (path === "/favicon.ico") return false;
  if (path === "/robots.txt") return false;
  if (path === "/sitemap.xml") return false;
  return true;
}

/**
 * Determine if a request path is for a static asset.
 */
export function isStaticAsset(path: string): boolean {
  const lowerPath = path.toLowerCase();
  return STATIC_EXTENSIONS.some((ext) => lowerPath.endsWith(ext));
}

/**
 * Extract a human-readable browser name from a user agent string.
 */
export function getBrowserName(userAgent: string): string {
  if (!userAgent || userAgent === "-") return "Unknown";
  if (isBot(userAgent)) return "Bot";

  // Order matters — check more specific patterns first
  if (/Edg\//i.test(userAgent)) return "Edge";
  if (/OPR\//i.test(userAgent) || /Opera/i.test(userAgent)) return "Opera";
  if (/Vivaldi/i.test(userAgent)) return "Vivaldi";
  if (/Brave/i.test(userAgent)) return "Brave";
  if (/Chrome/i.test(userAgent) && !/Chromium/i.test(userAgent)) return "Chrome";
  if (/Chromium/i.test(userAgent)) return "Chromium";
  if (/Firefox/i.test(userAgent)) return "Firefox";
  if (/Safari/i.test(userAgent) && !/Chrome/i.test(userAgent)) return "Safari";
  if (/MSIE|Trident/i.test(userAgent)) return "Internet Explorer";

  return "Other";
}

/**
 * Format a byte count into a human-readable string.
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const value = bytes / Math.pow(1024, i);
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

// ─── Log Parsing ──────────────────────────────────────────────────────────────

/**
 * Parse a single Apache Combined Log Format line into a LogEntry.
 * Returns null if the line cannot be parsed.
 */
export function parseLogLine(line: string): LogEntry | null {
  const match = line.match(LOG_LINE_REGEX);
  if (!match) return null;

  const ip = match[1];
  const timestampStr = match[2];
  const method = match[3];
  const path = match[4];
  const protocol = match[5];
  const statusStr = match[6];
  const bytesStr = match[7];
  const referer = match[8];
  const userAgent = match[9];

  if (!ip || !timestampStr || !method || !path || !protocol || !statusStr || !bytesStr || !referer || userAgent === undefined) {
    return null;
  }

  // Parse Apache timestamp: "10/Oct/2023:13:55:36 +0000"
  const timestamp = parseApacheTimestamp(timestampStr);
  if (!timestamp) return null;

  const status = parseInt(statusStr, 10);
  const bytes = bytesStr === "-" ? 0 : parseInt(bytesStr, 10);

  return {
    ip,
    timestamp,
    method,
    path,
    protocol,
    status,
    bytes,
    referer,
    userAgent,
  };
}

/**
 * Parse Apache timestamp format: "10/Oct/2023:13:55:36 +0000"
 */
function parseApacheTimestamp(str: string): Date | null {
  const months: Record<string, number> = {
    Jan: 0, Feb: 1, Mar: 2, Apr: 3, May: 4, Jun: 5,
    Jul: 6, Aug: 7, Sep: 8, Oct: 9, Nov: 10, Dec: 11,
  };

  const match = str.match(
    /(\d{2})\/(\w{3})\/(\d{4}):(\d{2}):(\d{2}):(\d{2}) ([+-]\d{4})/,
  );
  if (!match) return null;

  const day = match[1];
  const monthStr = match[2];
  const year = match[3];
  const hour = match[4];
  const minute = match[5];
  const second = match[6];
  const tz = match[7];

  if (!day || !monthStr || !year || !hour || !minute || !second || !tz) {
    return null;
  }

  const month = months[monthStr];
  if (month === undefined) return null;

  // Parse timezone offset
  const tzSign = tz[0] === "+" ? 1 : -1;
  const tzHours = parseInt(tz.slice(1, 3), 10);
  const tzMinutes = parseInt(tz.slice(3, 5), 10);
  const tzOffsetMs = tzSign * (tzHours * 60 + tzMinutes) * 60 * 1000;

  // Create date in UTC then adjust for timezone
  const date = new Date(
    Date.UTC(
      parseInt(year, 10),
      month,
      parseInt(day, 10),
      parseInt(hour, 10),
      parseInt(minute, 10),
      parseInt(second, 10),
    ),
  );
  date.setTime(date.getTime() - tzOffsetMs);

  return date;
}

/**
 * Parse Apache access log files (main log + rotated logs).
 * Reads the main log and up to 2 rotated logs (.1 and .2.gz).
 */
export async function parseLogFiles(
  basePath: string = LOG_PATH,
): Promise<LogEntry[]> {
  const entries: LogEntry[] = [];

  // Try main log file
  try {
    const content = await Deno.readTextFile(basePath);
    const lines = content.split("\n");
    for (const line of lines) {
      if (line.trim()) {
        const entry = parseLogLine(line);
        if (entry) entries.push(entry);
      }
    }
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      console.error(`Error reading log file ${basePath}:`, error);
    }
  }

  // Try rotated log .1 (uncompressed)
  try {
    const content = await Deno.readTextFile(`${basePath}.1`);
    const lines = content.split("\n");
    for (const line of lines) {
      if (line.trim()) {
        const entry = parseLogLine(line);
        if (entry) entries.push(entry);
      }
    }
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      console.error(`Error reading rotated log ${basePath}.1:`, error);
    }
  }

  // Try rotated log .2.gz (compressed)
  try {
    const compressedData = await Deno.readFile(`${basePath}.2.gz`);
    const ds = new DecompressionStream("gzip");
    const decompressedStream = new Blob([compressedData]).stream().pipeThrough(ds);
    const reader = decompressedStream.getReader();
    const chunks: Uint8Array[] = [];

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(value);
    }

    const decoder = new TextDecoder();
    const content = chunks.map((chunk) => decoder.decode(chunk, { stream: true })).join("") +
      decoder.decode();
    const lines = content.split("\n");
    for (const line of lines) {
      if (line.trim()) {
        const entry = parseLogLine(line);
        if (entry) entries.push(entry);
      }
    }
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      // .gz file may not exist, that's fine
      if (!(error instanceof TypeError)) {
        console.error(`Error reading compressed log ${basePath}.2.gz:`, error);
      }
    }
  }

  // Sort by timestamp (oldest first)
  entries.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime());

  return entries;
}

// ─── Statistics Generation ────────────────────────────────────────────────────

/**
 * Generate statistics from parsed log entries.
 *
 * @param entries - Parsed log entries
 * @param excludeOwner - If true, exclude the owner's IP from statistics
 * @param excludedHost - The owner's IP address to exclude (if excludeOwner is true)
 */
export function generateStatistics(
  entries: LogEntry[],
  excludeOwner = true,
  excludedHost?: string,
): Statistics {
  // Filter entries if excluding owner
  const filtered = excludeOwner && excludedHost
    ? entries.filter((e) => e.ip !== excludedHost)
    : entries;

  const uniqueIPs = new Set<string>();
  const pageCounts = new Map<string, number>();
  const ipCounts = new Map<string, number>();
  const statusCodes = new Map<number, number>();
  const requestsByHour = new Map<number, number>();
  const requestsByDay = new Map<string, number>();
  const browsers = new Map<string, number>();
  let pageViews = 0;
  let bandwidth = 0;
  let botRequests = 0;
  let humanRequests = 0;

  for (const entry of filtered) {
    // Track unique visitors
    uniqueIPs.add(entry.ip);

    // Track bandwidth
    bandwidth += entry.bytes;

    // Track bot vs human
    if (isBot(entry.userAgent)) {
      botRequests++;
    } else {
      humanRequests++;
    }

    // Track page views (only non-bot, non-static requests)
    if (!isBot(entry.userAgent) && isPageView(entry.path)) {
      pageViews++;
      const currentCount = pageCounts.get(entry.path) ?? 0;
      pageCounts.set(entry.path, currentCount + 1);
    }

    // Track IP counts
    const ipCount = ipCounts.get(entry.ip) ?? 0;
    ipCounts.set(entry.ip, ipCount + 1);

    // Track status codes
    const statusCount = statusCodes.get(entry.status) ?? 0;
    statusCodes.set(entry.status, statusCount + 1);

    // Track requests by hour
    const hour = entry.timestamp.getHours();
    const hourCount = requestsByHour.get(hour) ?? 0;
    requestsByHour.set(hour, hourCount + 1);

    // Track requests by day
    const dayKey = entry.timestamp.toISOString().split("T")[0] ?? "";
    const dayCount = requestsByDay.get(dayKey) ?? 0;
    requestsByDay.set(dayKey, dayCount + 1);

    // Track browsers (only for human visitors)
    if (!isBot(entry.userAgent)) {
      const browser = getBrowserName(entry.userAgent);
      const browserCount = browsers.get(browser) ?? 0;
      browsers.set(browser, browserCount + 1);
    }
  }

  // Sort top pages by count (descending), take top 20
  const topPages = [...pageCounts.entries()]
    .sort((a, b) => b[1] - a[1])
    .slice(0, 20)
    .map(([path, count]) => ({ path, count }));

  // Sort top IPs by count (descending), take top 20
  const topIPs = [...ipCounts.entries()]
    .sort((a, b) => b[1] - a[1])
    .slice(0, 20)
    .map(([ip, count]) => {
      // Check if this IP is predominantly a bot by looking at its user agents
      const ipEntries = filtered.filter((e) => e.ip === ip);
      const botCount = ipEntries.filter((e) => isBot(e.userAgent)).length;
      return { ip, count, isBot: botCount > ipEntries.length / 2 };
    });

  return {
    totalRequests: filtered.length,
    uniqueVisitors: uniqueIPs.size,
    pageViews,
    bandwidth,
    topPages,
    topIPs,
    statusCodes,
    requestsByHour,
    requestsByDay,
    browsers,
    botRequests,
    humanRequests,
  };
}
