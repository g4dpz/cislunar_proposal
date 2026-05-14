// email/mod.ts — SMTP email sending module using Deno's built-in net APIs

export interface EmailConfig {
  host: string;
  port: number;
  username: string;
  password: string;
  from: string;
  to: string;
  useTls: boolean;
}

export interface ContactMessage {
  name: string;
  callsignOrOrg: string;
  areaOfInterest: string;
  message: string;
}

/**
 * Load email configuration from environment variables.
 * Returns null if SMTP_HOST is not set (email disabled).
 */
export function loadEmailConfig(): EmailConfig | null {
  const host = Deno.env.get("SMTP_HOST");
  if (!host) {
    return null;
  }

  return {
    host,
    port: parseInt(Deno.env.get("SMTP_PORT") ?? "587", 10),
    username: Deno.env.get("SMTP_USERNAME") ?? "",
    password: Deno.env.get("SMTP_PASSWORD") ?? "",
    from: Deno.env.get("SMTP_FROM") ?? "noreply@cislunar-dtn.org",
    to: Deno.env.get("SMTP_TO") ?? "dave@g4dpz.me.uk",
    useTls: (Deno.env.get("SMTP_USE_TLS") ?? "true") === "true",
  };
}

/**
 * Send a contact form notification email via SMTP.
 * Uses STARTTLS on port 587 or direct TLS on port 465.
 */
export async function sendContactEmail(
  config: EmailConfig,
  contact: ContactMessage,
): Promise<void> {
  const subject = `[Cislunar DTN] Contact from ${contact.name}`;
  const body = buildEmailBody(contact);
  const message = buildMimeMessage(config.from, config.to, subject, body);

  const conn = config.useTls && config.port === 465
    ? await Deno.connectTls({ hostname: config.host, port: config.port })
    : await Deno.connect({ hostname: config.host, port: config.port });

  try {
    const encoder = new TextEncoder();
    const decoder = new TextDecoder();

    // Read greeting
    await readResponse(conn, decoder);

    // EHLO
    await writeCommand(conn, encoder, `EHLO cislunar-dtn.org\r\n`);
    await readResponse(conn, decoder);

    // STARTTLS if port 587
    let secureConn: Deno.TlsConn | Deno.Conn = conn;
    if (config.useTls && config.port !== 465) {
      await writeCommand(conn, encoder, `STARTTLS\r\n`);
      await readResponse(conn, decoder);
      secureConn = await Deno.startTls(conn as Deno.TcpConn, {
        hostname: config.host,
      });
      // Re-EHLO after STARTTLS
      await writeCommand(secureConn, encoder, `EHLO cislunar-dtn.org\r\n`);
      await readResponse(secureConn, decoder);
    }

    // AUTH LOGIN
    await writeCommand(secureConn, encoder, `AUTH LOGIN\r\n`);
    await readResponse(secureConn, decoder);

    await writeCommand(secureConn, encoder, `${btoa(config.username)}\r\n`);
    await readResponse(secureConn, decoder);

    await writeCommand(secureConn, encoder, `${btoa(config.password)}\r\n`);
    await readResponse(secureConn, decoder);

    // MAIL FROM
    await writeCommand(
      secureConn,
      encoder,
      `MAIL FROM:<${config.from}>\r\n`,
    );
    await readResponse(secureConn, decoder);

    // RCPT TO
    await writeCommand(secureConn, encoder, `RCPT TO:<${config.to}>\r\n`);
    await readResponse(secureConn, decoder);

    // DATA
    await writeCommand(secureConn, encoder, `DATA\r\n`);
    await readResponse(secureConn, decoder);

    // Message content
    await writeCommand(secureConn, encoder, `${message}\r\n.\r\n`);
    await readResponse(secureConn, decoder);

    // QUIT
    await writeCommand(secureConn, encoder, `QUIT\r\n`);
    await readResponse(secureConn, decoder);
  } finally {
    try {
      conn.close();
    } catch { /* ignore close errors */ }
  }
}

function buildEmailBody(contact: ContactMessage): string {
  const lines = [
    `New contact form submission from the Cislunar DTN website:`,
    ``,
    `Name: ${contact.name}`,
  ];

  if (contact.callsignOrOrg) {
    lines.push(`Callsign/Organisation: ${contact.callsignOrOrg}`);
  }
  if (contact.areaOfInterest) {
    lines.push(`Area of Interest: ${contact.areaOfInterest}`);
  }

  lines.push(``, `Message:`, `${contact.message}`);

  return lines.join("\r\n");
}

function buildMimeMessage(
  from: string,
  to: string,
  subject: string,
  body: string,
): string {
  const date = new Date().toUTCString();
  return [
    `From: ${from}`,
    `To: ${to}`,
    `Subject: ${subject}`,
    `Date: ${date}`,
    `MIME-Version: 1.0`,
    `Content-Type: text/plain; charset=UTF-8`,
    ``,
    body,
  ].join("\r\n");
}

async function writeCommand(
  conn: Deno.Conn,
  encoder: TextEncoder,
  command: string,
): Promise<void> {
  await conn.write(encoder.encode(command));
}

async function readResponse(
  conn: Deno.Conn,
  decoder: TextDecoder,
): Promise<string> {
  const buf = new Uint8Array(1024);
  const n = await conn.read(buf);
  if (n === null) throw new Error("SMTP connection closed unexpectedly");
  const response = decoder.decode(buf.subarray(0, n));
  const code = parseInt(response.substring(0, 3), 10);
  if (code >= 400) {
    throw new Error(`SMTP error: ${response.trim()}`);
  }
  return response;
}
