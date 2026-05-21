// content/data.ts — Static site content authored from README.md reference

// ─── Interfaces ───────────────────────────────────────────────────────────────

export interface ResourceLink {
  title: string;
  url: string;
  description: string;
}

export interface Collaborator {
  name: string;
  logoUrl: string;
  websiteUrl: string;
}

export interface NavItem {
  label: string;
  href: string;
  active: boolean;
}

export interface ContactFormField {
  name: string;
  label: string;
  type: "text" | "email" | "textarea" | "select";
  required: boolean;
  options?: string[];
}

export interface RoadmapPhase {
  id: string;
  name: string;
  status: "complete" | "in-progress" | "planned";
  hardware: string;
  link: string;
  purpose: string;
  docsPath: string;
}

export interface OverviewContent {
  title: string;
  tagline: string;
  missionSummary: string;
  missionSummary2: string;
  features: string[];
  protocolStack: string[];
  license: string;
  collaborators: Collaborator[];
}

export interface ContactContent {
  email: string;
  callsigns: string[];
  collaborators: Collaborator[];
  targetGroups: string[];
  formFields: ContactFormField[];
}

export interface PrivacyContent {
  dataController: string;
  contactEmail: string;
  dataCollected: string;
  legalBasis: string;
  retentionPeriod: string;
  cookiePolicy: string;
}

export interface ConOpsContent {
  concept: string;
  rfLinkTypes: string[];
  whyCommunityMatters: string[];
  expectedOutcomes: string[];
  nasaReferences: ResourceLink[];
}

export interface DocumentationLinks {
  phases: { name: string; docsPath: string }[];
  externalRefs: ResourceLink[];
  packages: ResourceLink[];
}

export interface PageData {
  meta: {
    title: string;
    description: string;
    keywords: string[];
    ogTitle: string;
    ogDescription: string;
    ogImage?: string;
    canonicalUrl: string;
  };
  nav: NavItem[];
  activeSection: string;
  content: Record<string, unknown>;
  collaborators: Collaborator[];
  currentYear: number;
  user?: {
    id: number;
    name: string;
    email: string;
    roles: Array<{ id: number; name: string; description: string }>;
    isAdmin?: boolean;
  } | null;
}

export interface SiteContent {
  overview: OverviewContent;
  roadmap: RoadmapPhase[];
  documentation: DocumentationLinks;
  contact: ContactContent;
  privacy: PrivacyContent;
  conops: ConOpsContent;
  nav: NavItem[];
}

// ─── Content Data ─────────────────────────────────────────────────────────────

const collaborators: Collaborator[] = [
  {
    name: "AMSAT-UK",
    logoUrl: "/images/logos/amsat-uk.png",
    websiteUrl: "https://amsat-uk.org",
  },
  {
    name: "AMSAT-DL",
    logoUrl: "/images/logos/amsat-dl.png",
    websiteUrl: "https://amsat-dl.org",
  },
  {
    name: "Goonhilly Earth Station",
    logoUrl: "/images/logos/goonhilly.png",
    websiteUrl: "https://goonhilly.org",
  },
];


const overview: OverviewContent = {
  title: "RADIANT — Radio Amateur Delay-tolerant Interplanetary Networking Testbed",
  tagline: "From amateur packet radio to CubeSat relay to cislunar networking.",
  missionSummary:
    "RADIANT brings Delay-Tolerant Networking (DTN) to amateur radio, enabling " +
    "store-and-forward messaging across disrupted links from terrestrial ground stations " +
    "to Low Earth Orbit (LEO) and ultimately to cislunar space.",
  missionSummary2:
    "Built on NASA Glenn's HDTN (High-rate Delay Tolerant Networking), this project implements " +
    "the Bundle Protocol version 7 (BPv7) over amateur radio links using LTP wrapped directly " +
    "in KISS framing. Station identification is achieved through callsign-embedded DTN Endpoint " +
    "Identifiers (e.g. dtn://g4dpz/spacecraft) carried in every bundle's metadata, ensuring " +
    "regulatory compliance while using numeric ipn:// addresses for efficient routing.",
  features: [
    "Working 3-node cislunar simulation with true packet-level propagation delay",
    "Demonstrated Earth-Moon (1.3s) and Earth-Mars (3-12 min) DTN store-and-forward",
    "LTP-over-KISS with callsign-embedded DTN Endpoint Identifiers (amateur radio compliance)",
    "Contact Graph Routing (CGR) computing multi-hop paths through relay nodes",
    "No encryption or cryptography (amateur radio regulatory compliance)",
    "Priority-based bundle handling (critical, expedited, normal, bulk)",
    "Persistent bundle storage surviving power cycles",
    "Real-time telemetry and health monitoring via HDTN REST API",
  ],
  protocolStack: [
    "Application (bping, bpsendfile)",
    "BPv7 (Bundle Protocol) — EID: dtn://callsign/service",
    "LTP (Licklider Transmission)",
    "KISS (TNC Serial Framing)",
    "USB Serial (TNC4)",
    "G3RUH GFSK (9600 baud)",
  ],
  license: "MIT",
  collaborators,
};

const roadmap: RoadmapPhase[] = [
  {
    id: "phase-1",
    name: "Phase 1: Terrestrial DTN Validation",
    status: "in-progress",
    hardware: "Raspberry Pi + Mobilinkd TNC4 + Yaesu FT-817",
    link: "VHF/UHF at 9600 baud (G3RUH GFSK)",
    purpose:
      "Ground-based validation using commercial amateur radio equipment. " +
      "Two-node terrestrial network operational. Validated ping, store-and-forward, " +
      "and telemetry.",
    docsPath: "docs/terrestrial-dtn-phase1/",
  },
  {
    id: "phase-1-5",
    name: "Phase 1.5: QO-100 GEO Satellite DTN",
    status: "planned",
    hardware: "Ground station with 2.4 GHz uplink + 10 GHz downlink",
    link: "QO-100 narrowband transponder (Es'hail-2 satellite)",
    purpose:
      "Geostationary satellite demonstration using QO-100 amateur transponder. " +
      "Validate DTN over real satellite link with constant visibility. " +
      "First space-based DTN demonstration before LEO orbital complexity.",
    docsPath: "docs/qo-100-geo-satellite-dtn/",
  },
  {
    id: "phase-2",
    name: "Phase 2: CubeSat Engineering Model (EM)",
    status: "planned",
    hardware: "STM32U585 OBC + Ettus B200mini SDR + External NVM",
    link: "UHF 437 MHz / S-band 2.2 GHz",
    purpose:
      "Ground-based flatsat with flight-representative hardware. " +
      "Validate flight software, power budget, thermal/vacuum readiness. " +
      "Identical software stack to flight unit, lab-grade RF front-end.",
    docsPath: "docs/cubesat-em-phase2/",
  },
  {
    id: "phase-3",
    name: "Phase 3: LEO CubeSat Flight",
    status: "planned",
    hardware: "STM32U585 OBC + Flight IQ transceiver",
    link: "UHF 437 MHz at 9.6 kbps",
    purpose:
      "Orbital deployment demonstrating ground-to-space DTN. " +
      "Ground-to-space ping and store-and-forward. " +
      "Handheld/small Yagi reception for broad community participation.",
    docsPath: "docs/leo-cubesat-phase3/",
  },
  {
    id: "phase-4",
    name: "Phase 4: Cislunar Deep-Space Communication",
    status: "planned",
    hardware: "STM32U585 or more capable processor",
    link: "S-band/X-band with LDPC/Turbo coding",
    purpose:
      "Amateur participation in Earth-Moon DTN. " +
      "Earth-Moon distance (~384,400 km). " +
      "3-5m dishes for 500 bps cislunar links. " +
      "Seeking support from the ESA ARTES programme for the prospective cislunar payload.",
    docsPath: "docs/cislunar-phase4/",
  },
];


const documentation: DocumentationLinks = {
  phases: [
    { name: "Phase 1: Terrestrial DTN", docsPath: "/docs/terrestrial-dtn-phase1" },
    { name: "Phase 1.5: QO-100 GEO Satellite DTN", docsPath: "/docs/qo-100-geo-satellite-dtn" },
    { name: "Phase 2: Engineering Model", docsPath: "/docs/cubesat-em-phase2" },
    { name: "Phase 3: LEO CubeSat", docsPath: "/docs/leo-cubesat-phase3" },
    { name: "Phase 4: Cislunar", docsPath: "/docs/cislunar-phase4" },
  ],
  externalRefs: [
    {
      title: "NASA Glenn: High-Rate Delay Tolerant Networking",
      url: "https://www.nasa.gov/glenn/glenn-expertise-space-exploration/scan/high-rate-delay-tolerant-networking/",
      description: "NASA Glenn Research Center's HDTN programme — the foundation software stack used by RADIANT.",
    },
    {
      title: "RFC 9171: Bundle Protocol Version 7",
      url: "https://www.rfc-editor.org/rfc/rfc9171.html",
      description: "The core Bundle Protocol specification for DTN.",
    },
    {
      title: "RFC 5326: Licklider Transmission Protocol (LTP)",
      url: "https://www.rfc-editor.org/rfc/rfc5326.html",
      description: "Reliable transmission protocol designed for deep-space links.",
    },
    {
      title: "KISS Protocol Specification",
      url: "http://www.ax25.net/kiss.aspx",
      description: "TNC serial framing protocol used for LTP segment transport.",
    },
  ],
  packages: [
    {
      title: "HDTN Wrapper",
      url: "https://github.com/g4dpz/cislunar_proposal/tree/main/pkg/hdtn",
      description: "Go wrapper for NASA Glenn's HDTN library.",
    },
    {
      title: "Contact Plan Manager + CGR",
      url: "https://github.com/g4dpz/cislunar_proposal/tree/main/pkg/contact",
      description: "Contact plan management and Contact Graph Routing.",
    },
    {
      title: "Security Package",
      url: "https://github.com/g4dpz/cislunar_proposal/tree/main/pkg/security",
      description: "Rate limiting and access control.",
    },
  ],
};

const contact: ContactContent = {
  email: "dave@g4dpz.me.uk",
  callsigns: [],
  collaborators,
  targetGroups: [
    "AMSAT organisations",
    "Amateur radio clubs",
    "Packet radio operators",
    "Microwave experimenters",
    "EME / weak-signal operators",
    "CubeSat teams",
    "Universities",
    "Space networking researchers",
  ],
  formFields: [
    {
      name: "email",
      label: "Email",
      type: "email",
      required: true,
    },
    {
      name: "callsign_or_org",
      label: "Callsign / Organisation",
      type: "text",
      required: false,
    },
    {
      name: "area_of_interest",
      label: "Area of Interest",
      type: "select",
      required: true,
      options: [
        "Ground station partnership",
        "Hardware contribution",
        "Software development",
        "RF link optimisation",
        "Orbital mechanics",
        "Documentation",
        "Testing",
        "Flight opportunity",
        "General enquiry",
      ],
    },
    {
      name: "message",
      label: "Message",
      type: "textarea",
      required: true,
    },
  ],
};

const privacy: PrivacyContent = {
  dataController:
    "Cislunar Amateur DTN Payload Project Team",
  contactEmail: "secretary@amsat-uk.org",
  dataCollected:
    "When you submit the contact form, we collect your name, callsign or organisation, " +
    "area of interest, and message content. We do not use cookies or third-party analytics " +
    "on this website. Server access logs may record your IP address and browser user-agent " +
    "for security and operational purposes.",
  legalBasis:
    "We process contact form submissions on the basis of legitimate interest (responding " +
    "to your enquiry) and, where applicable, your consent. Server logs are processed on " +
    "the basis of legitimate interest in maintaining website security.",
  retentionPeriod:
    "Contact form submissions are retained for up to 24 months or until your enquiry is " +
    "resolved, whichever is longer. Server access logs are retained for up to 90 days.",
  cookiePolicy:
    "This website does not use cookies or third-party tracking. No analytics services " +
    "are employed. If this changes in the future, this page will be updated and consent " +
    "will be obtained where required.",
};


const conops: ConOpsContent = {
  concept:
    "The terrestrial analogue architecture mirrors the cislunar communications path: " +
    "Mission Operations Node → Ground/Gateway DTN Node → Amateur/experimental RF link → " +
    "Relay/Remote DTN Node → Payload application endpoint. Each terrestrial demonstration " +
    "exercises the same DTN protocols, store-and-forward behaviour, and contact scheduling " +
    "that a cislunar mission would require, using amateur radio links as realistic " +
    "disrupted-network analogues.",
  rfLinkTypes: [
    "Amateur radio LTP-over-KISS (VHF/UHF at 9600 baud, callsign in DTN EID)",
    "Microwave point-to-point paths (10 GHz, 2.4 GHz)",
    "Satellite-style scheduled links (QO-100 transponder)",
    "EME-inspired operational patterns (moonbounce timing disciplines)",
    "Weak-signal / intermittent links (troposcatter, aircraft scatter)",
  ],
  whyCommunityMatters: [
    "Operation over difficult and intermittent RF paths",
    "Disciplined link scheduling and station procedures",
    "Experience with weak-signal and long-distance communication",
    "Practical ground-segment engineering and experimentation",
    "Collaborative culture suited to distributed demonstrations",
  ],
  expectedOutcomes: [
    "A credible terrestrial analogue for cislunar DTN operations",
    "Operational evidence to support a flight experiment case",
    "Stronger links between space networking and specialist amateur-radio communities",
    "A clearer roadmap toward a cislunar payload demonstration of DTN-enabled networking",
  ],
  nasaReferences: [
    {
      title: "NASA Glenn: High-Rate Delay Tolerant Networking",
      url: "https://www.nasa.gov/glenn/glenn-expertise-space-exploration/scan/high-rate-delay-tolerant-networking/",
      description: "NASA Glenn Research Center's HDTN programme page — architecture overview, capabilities, and mission context.",
    },
    {
      title: "Delay/Disruption Tolerant Networking Tutorial",
      url: "https://www.nasa.gov/dtn-tutorial",
      description: "NASA educational resource introducing DTN concepts and architecture.",
    },
    {
      title: "High-rate Delay Tolerant Networking (HDTN)",
      url: "https://github.com/nasa/HDTN",
      description: "NASA Glenn's high-performance DTN software suite.",
    },
    {
      title: "Delay/Disruption Tolerant Networking Overview",
      url: "https://www.nasa.gov/dtn-overview",
      description: "High-level overview of NASA's DTN programme.",
    },
    {
      title: "NASA Delay Tolerant Networks: Operational, Evolving, and Ready for Expansion",
      url: "https://www.nasa.gov/dtn-expansion",
      description: "Status report on NASA's operational DTN deployment.",
    },
    {
      title: "A Communications Network for Cislunar Operations",
      url: "https://www.nasa.gov/cislunar-comms",
      description: "NASA architecture for cislunar communication networks.",
    },
  ],
};

const nav: NavItem[] = [
  { label: "Home", href: "/", active: false },
  { label: "Roadmap", href: "/roadmap", active: false },
  { label: "ConOps", href: "/conops", active: false },
  { label: "Documentation", href: "/docs", active: false },
  { label: "Contact", href: "/contact", active: false },
  { label: "Privacy", href: "/privacy", active: false },
];

// ─── Exported Site Content ────────────────────────────────────────────────────

export const siteContent: SiteContent = {
  overview,
  roadmap,
  documentation,
  contact,
  privacy,
  conops,
  nav,
};
