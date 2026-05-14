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

export interface ResourcesContent {
  categories: { name: string; links: ResourceLink[] }[];
}

export interface ConOpsContent {
  concept: string;
  rfLinkTypes: string[];
  whyCommunityMatters: string[];
  expectedOutcomes: string[];
  nasaReferences: ResourceLink[];
}

export interface GettingStartedContent {
  prerequisites: string[];
  installation: string[];
  runNetwork: string[];
  testConnectivity: string[];
}

export interface ContributingContent {
  areas: string[];
  developmentSetup: string[];
  license: string;
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
}

export interface SiteContent {
  overview: OverviewContent;
  roadmap: RoadmapPhase[];
  gettingStarted: GettingStartedContent;
  contributing: ContributingContent;
  documentation: DocumentationLinks;
  contact: ContactContent;
  privacy: PrivacyContent;
  resources: ResourcesContent;
  conops: ConOpsContent;
  nav: NavItem[];
}

// ─── Content Data ─────────────────────────────────────────────────────────────

const collaborators: Collaborator[] = [
  {
    name: "AMSAT-UK",
    logoUrl: "https://amsat-uk.org/media/press/amsat-uk_bevelled_logo_with_title/",
    websiteUrl: "https://amsat-uk.org",
  },
  {
    name: "AMSAT-DL",
    logoUrl: "/images/amsat-dl-logo.png",
    websiteUrl: "https://amsat-dl.org",
  },
];


const overview: OverviewContent = {
  title: "Amateur Radio DTN Space Networking Pathfinder",
  tagline: "From amateur packet radio to CubeSat relay to cislunar networking.",
  missionSummary:
    "The Cislunar Amateur DTN Payload project brings Delay-Tolerant Networking (DTN) " +
    "to amateur radio, enabling store-and-forward messaging across disrupted links from " +
    "terrestrial ground stations to Low Earth Orbit (LEO) and ultimately to cislunar space. " +
    "Built on NASA Glenn's HDTN (High-rate Delay Tolerant Networking), this project implements " +
    "the Bundle Protocol version 7 (BPv7) over amateur radio links.",
  features: [
    "AX.25 link-layer framing with callsign addressing (amateur radio compliance)",
    "No encryption or cryptography (amateur radio regulatory compliance)",
    "Automated orbital pass prediction using Contact Graph Routing (CGR)",
    "Priority-based bundle handling (critical, expedited, normal, bulk)",
    "Persistent bundle storage surviving power cycles",
    "Real-time telemetry and health monitoring",
  ],
  protocolStack: [
    "Application (bping, bpsendfile)",
    "BPv7 (Bundle Protocol)",
    "LTP (Licklider Transmission)",
    "AX.25 (Amateur Radio Link Layer)",
    "KISS (TNC Serial Protocol)",
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
    status: "complete",
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
      "3-5m dishes for 500 bps cislunar links.",
    docsPath: "docs/cislunar-phase4/",
  },
];

const gettingStarted: GettingStartedContent = {
  prerequisites: [
    "Linux or macOS (amd64/arm64)",
    "Go 1.19 or later",
    "Two Mobilinkd TNC4 terminal node controllers (USB)",
    "Two Yaesu FT-817 radios configured for 9600 baud",
    "Amateur radio license (required for transmission)",
  ],
  installation: [
    "git clone https://github.com/g4dpz/cislunar_proposal.git",
    "cd cislunar_proposal",
    "cd HDTN && mkdir build && cd build",
    "cmake .. -DCMAKE_INSTALL_PREFIX=$(pwd)/../../hdtn-install",
    "make -j$(nproc) && make install",
    "cd ../..",
    "go build -o dtn-node ./cmd/dtn-node",
  ],
  runNetwork: [
    "./dtn-node -config configs/dtn-node-a.yaml",
    "./dtn-node -config configs/dtn-node-b.yaml",
  ],
  testConnectivity: [
    "export PATH=$PATH:$(pwd)/hdtn-install/bin",
    "bping ipn:1.1 ipn:2.1 -c 5",
    "bpsendfile ipn:1.1 ipn:2.1 test-message.txt",
  ],
};

const contributing: ContributingContent = {
  areas: [
    "Bug reports and fixes",
    "Documentation improvements",
    "Additional test coverage",
    "Hardware integration (TNC4, B200mini, STM32U585)",
    "RF link optimization",
    "Orbital mechanics improvements",
  ],
  developmentSetup: [
    "git clone https://github.com/g4dpz/cislunar_proposal.git",
    "cd cislunar_proposal",
    "go mod download",
    "go test ./...",
    "go build -o dtn-node ./cmd/dtn-node",
    "go build -o em-node ./cmd/em-node",
    "go build -o leo-node ./cmd/leo-node",
    "go build -o cislunar-node ./cmd/cislunar-node",
  ],
  license: "MIT",
};


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
      title: "AX.25 Link Access Protocol",
      url: "http://www.ax25.net/AX25.2.2-Jul%2098-2.pdf",
      description: "Amateur radio data link layer protocol specification.",
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
  callsigns: ["M0DTN"],
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
      name: "name",
      label: "Name",
      type: "text",
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
  contactEmail: "privacy@cislunar-dtn.org",
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

const resources: ResourcesContent = {
  categories: [
    {
      name: "NASA DTN Overview",
      links: [
        {
          title: "NASA Delay-Tolerant Networking Overview",
          url: "https://www.nasa.gov/directorates/heo/scan/engineering/technology/disruption_tolerant_networking",
          description:
            "Introduction to NASA's DTN programme and its applications in space communications.",
        },
        {
          title: "NASA DTN: Operational, Evolving, and Ready for Expansion",
          url: "https://www.nasa.gov/dtn-operational",
          description:
            "Overview of NASA's operational DTN deployment and future expansion plans.",
        },
      ],
    },
    {
      name: "NASA HDTN",
      links: [
        {
          title: "HDTN (High-rate Delay Tolerant Networking)",
          url: "https://github.com/nasa/HDTN",
          description:
            "NASA Glenn's high-performance DTN implementation designed for high data rates and modern architectures.",
        },
        {
          title: "HDTN Documentation",
          url: "https://github.com/nasa/HDTN/wiki",
          description:
            "Documentation and resources for the HDTN software suite.",
        },
      ],
    },
    {
      name: "Bundle Protocol References",
      links: [
        {
          title: "RFC 9171: Bundle Protocol Version 7",
          url: "https://www.rfc-editor.org/rfc/rfc9171.html",
          description: "The core Bundle Protocol specification for DTN.",
        },
        {
          title: "RFC 5326: Licklider Transmission Protocol (LTP)",
          url: "https://www.rfc-editor.org/rfc/rfc5326.html",
          description:
            "Reliable transmission protocol designed for deep-space links.",
        },
      ],
    },
    {
      name: "AMSAT Resources",
      links: [
        {
          title: "AMSAT-UK",
          url: "https://amsat-uk.org",
          description:
            "Radio Amateur Satellite Corporation of the United Kingdom.",
        },
        {
          title: "AMSAT-DL",
          url: "https://amsat-dl.org",
          description:
            "Radio Amateur Satellite Corporation of Germany.",
        },
      ],
    },
    {
      name: "AX.25 / KISS Background",
      links: [
        {
          title: "AX.25 Link Access Protocol Specification",
          url: "http://www.ax25.net/AX25.2.2-Jul%2098-2.pdf",
          description:
            "The AX.25 amateur radio data link layer protocol specification.",
        },
        {
          title: "KISS TNC Protocol",
          url: "https://www.ax25.net/kiss.aspx",
          description:
            "Keep It Simple, Stupid — the serial protocol for TNC communication.",
        },
      ],
    },
  ],
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
    "Amateur packet radio (VHF/UHF AX.25 at 9600 baud)",
    "Microwave point-to-point paths (10 GHz, 24 GHz)",
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
    {
      title: "NASA's Lunar Communications and Navigation Architecture",
      url: "https://www.nasa.gov/lunar-comms-nav",
      description: "The planned communications and navigation infrastructure for lunar missions.",
    },
  ],
};

const nav: NavItem[] = [
  { label: "Home", href: "/", active: false },
  { label: "Roadmap", href: "/roadmap", active: false },
  { label: "ConOps", href: "/conops", active: false },
  { label: "Documentation", href: "/docs", active: false },
  { label: "Resources", href: "/resources", active: false },
  { label: "Getting Started", href: "/getting-started", active: false },
  { label: "Contributing", href: "/contributing", active: false },
  { label: "Contact", href: "/contact", active: false },
  { label: "Privacy", href: "/privacy", active: false },
];

// ─── Exported Site Content ────────────────────────────────────────────────────

export const siteContent: SiteContent = {
  overview,
  roadmap,
  gettingStarted,
  contributing,
  documentation,
  contact,
  privacy,
  resources,
  conops,
  nav,
};
