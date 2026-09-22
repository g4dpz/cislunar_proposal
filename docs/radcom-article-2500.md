# RADIANT: Radio Amateur Delay-tolerant Interplanetary Networking Testbed

**David Johnson, G4DPZ**

*An open-source project building the amateur radio network that works when the internet can't — from your back garden to the Moon*

---

## A Familiar Problem in an Unfamiliar Place

Imagine setting up a ground station pointed at the Moon. You type a message, press send, and watch as your data travels at light speed toward an amateur radio payload in cislunar space. About 1.3 seconds later, it arrives. Another 1.3 seconds later, an acknowledgement returns. Your message has been stored onboard, waiting for the next contact window with a ground station in Europe.

This is the goal of RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — and we are building it today.

RADIANT is an open-source project, supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station, implementing the same protocols that NASA and ESA deploy for deep-space communications. The project adapts these for amateur operators using equipment from a Raspberry Pi and Yaesu FT-817 to dishes capable of receiving lunar signals.

---

## Why the Internet Does Not Work in Space

Most of us take the internet for granted. When browsing a website, your computer opens a connection, they exchange handshakes, and data flows almost instantaneously. This works because the internet assumes connections are fast, reliable, and continuous.

Space is none of those things.

London to New York is 30 milliseconds. A geostationary satellite adds 120 ms. The Moon is 1.3 seconds. Mars ranges from three to twelve minutes. Jupiter is 35 to 50 minutes.

Consider TCP/IP communication with the Moon. The TCP handshake alone requires three messages between Earth and Moon. At 1.3 seconds each way, that's nearly four seconds just to establish a connection. For Mars at average distance, a TCP handshake could take over an hour.

The problem isn't bandwidth — modern radio links achieve reasonable data rates over enormous distances. The problem is latency and intermittency. A LEO satellite is visible for perhaps ten minutes per pass. A cislunar spacecraft may have its antenna pointed away from Earth for hours. The continuous connections we assume on Earth simply don't exist in space.

---

## Delay-Tolerant Networking: A Postal Service for Data

The solution is Delay-Tolerant Networking (DTN). Rather than assuming continuous connectivity, DTN works like a postal service. Data is packaged into "bundles" which are stored at each network node and forwarded whenever communication opportunities arise.

If the next hop isn't available — perhaps the satellite hasn't risen above the horizon — the node holds the bundle until the path opens. When the bundle reaches its destination, an acknowledgement returns through the chain.

**[IMAGE: Diagram showing LTP store-and-forward flow with Earth station, satellite, and destination. Shows data being stored at intermediate nodes when links are unavailable. Caption: "LTP store-and-forward operation: data waits at each node until the next communication opportunity becomes available."]**

This store-and-forward model will be familiar to anyone who remembers packet BBS operations in the 1980s. Messages were stored at each node and forwarded when propagation or scheduled links allowed. DTN formalises this using modern protocols standardised by the IETF and CCSDS.

The core protocols are:

- **Bundle Protocol version 7 (BPv7)** — creates, stores, and delivers bundles
- **Licklider Transmission Protocol (LTP)** — provides reliable delivery with deferred acknowledgement for long round-trip delays  
- **Contact Graph Routing (CGR)** — computes paths through time-varying networks using predicted contact windows

These aren't experimental. NASA has developed DTN since the early 2000s. The International Space Station runs DTN experiments. Korea's lunar orbiter tested Bundle Protocol at lunar distance in 2024. ESA's Moonlight programme uses these protocols as its foundation.

The protocols RADIANT uses are becoming the standard for space communications. In 2023, the Internet Society's Interplanetary Chapter published a Solar System Internet governance framework co-authored by Vint Cerf, widely recognized as one of the "fathers of the Internet" for creating TCP/IP. Their recommendation: DTN and Bundle Protocol should be the foundation for all space communications that exceed IP's capabilities.

Both ESA's Moonlight programme and NASA's LunaNet specification mandate these same protocols. Through RADIANT, amateur radio participates as a stakeholder in this emerging interplanetary infrastructure.

---

## How RADIANT Sends Data Over Amateur Radio

RADIANT uses familiar KISS framing — the same approach TNCs have used for decades. The difference is what sits inside those frames.

Instead of AX.25 packets, RADIANT carries DTN data directly in KISS frames, eliminating AX.25 entirely — 15 bytes saved per frame (~10% improvement). Messages are packaged as "bundles" using Bundle Protocol. LTP handles reliable delivery over long-delay links. KISS feeds the data to your TNC and radio as normal.

The architecture uses ION-DTN as the primary implementation, ensuring compatibility with NASA's standard DTN stack used across professional space missions.

---

## Station Identification: Callsigns in the Protocol

RADIANT solves amateur radio's station identification requirement by embedding callsigns directly into DTN Endpoint Identifiers. Every bundle carries a source address like `dtn://g4dpz/service`, meaning the operator's callsign appears in every transmission.

Additionally, the system transmits a plaintext beacon every ten minutes containing station callsign, grid locator, and node type. Any station that can demodulate the signal can identify the transmitter.

There is no encryption anywhere in the system. All data travels in the clear, keeping RADIANT fully compliant with ITU Radio Regulations and national amateur radio rules.

---

## The Four-Phase Roadmap

RADIANT follows a phased approach, with each phase validating technology needed for the next.

### Phase 1: Terrestrial Validation (Complete)

Uses equipment familiar to packet operators: Raspberry Pi, Mobilinkd TNC4, and Yaesu FT-817 at 9600 baud BPSK. This validates the complete software stack over real amateur radio links before committing to space hardware.

**[IMAGE: Photo of RADIANT Phase 1 test setup showing Yaesu FT-817, Mobilinkd TNC4, and Raspberry Pi. Caption: "RADIANT Phase 1 terrestrial testing uses familiar amateur packet radio equipment."]**

### Phase 2: QO-100 (In Progress)  

Es'hail-2 provides an ideal space testbed with its always-on transponder and 250ms delays. Technical specifications:
- 1200 baud QPSK modulation
- Standard QO-100 ground station equipment

**[IMAGE: Photo of QO-100 ground station setup with dish, feeds, and transceiver equipment. Caption: "RADIANT Phase 2 uses standard QO-100 ground station equipment - familiar to many amateur operators already active on Es'hail-2."]**

**[IMAGE: Photo of Ettus B200 QPSK test bench setup. Caption: "RADIANT QPSK development and testing using Ettus B200 SDR platform."]**

This validates LTP's deferred acknowledgement over a real space link — the critical stepping stone between ground-based delays and cislunar 1.3-second delays.

### Phase 3: LEO CubeSat (Planned)

The first RADIANT node in orbit operating on 437 MHz UHF at 4.8 kbps QPSK. Technical specifications:
- STM32U585 microcontroller (160 MHz, 2 MB flash)
- 31 dB link budget margin with standard UHF Yagi
- 64-256 MB persistent storage
- Compatible with existing amateur satellite equipment

This demonstrates ground-to-space DTN messaging and orbital store-and-forward.

### Phase 4: Cislunar (Planned)

An amateur DTN node between Earth and Moon. Technical specifications:
- S-band 2.2 GHz at 500 bps QPSK, 5W + 10 dBi spacecraft antenna
- 35 dBi ground dish (3-5m, EME-class), 7 dB link margin
- 2.6-second lunar round-trip delay handling

This would be the first amateur-operated interplanetary-style communication system.

---

## What We Have Already Achieved

RADIANT operates a three-node cislunar simulation with true 1.3-second Earth-Moon delays. Contact Graph Routing computes multi-hop paths while LTP manages extended round-trips. The system successfully manages deferred acknowledgements and delivers data despite delays.

The project uses ION-DTN as the primary implementation, validating compatibility with NASA's standard DTN stack. The entire codebase is open-source under MIT licence with automated testing.

---

## Ground Station Requirements

RADIANT emphasizes accessibility:

**Phase 1:** Any 9600-baud packet setup — Raspberry Pi, TNC, and VHF/UHF radio.

**Phase 2:** Standard QO-100 ground station — 60-90cm dish, SDR, and appropriate feeds.

**Phase 3:** UHF Yagi antenna for 435 MHz amateur satellite work.

**Phase 4:** Dish antenna with low-noise amplification on S-band, comparable to EME station capabilities.

The software is open-source and runs on standard computing hardware.

---

## The Bigger Picture: A Solar System Internet

RADIANT isn't happening in isolation. The Internet Society's report envisions a "network of networks" in space, with multistakeholder governance meaning decisions about space networking shouldn't be made solely by space agencies. Amateur radio has a legitimate role.

**[IMAGE: NASA/ESA Solar System Internet concept diagram showing DTN nodes connecting Earth, Moon, Mars, and other destinations. Caption: "NASA and ESA vision for a Solar System Internet using DTN protocols - RADIANT demonstrates amateur radio's role in this emerging infrastructure."]**

This isn't unprecedented. Amateur operators experimented with packet networking before the commercial internet existed, and decades before commercial satellite constellations. Amateur radio pioneers technology that later becomes mainstream.

RADIANT is how we earn our place in the emerging Solar System Internet. By implementing the same protocols and demonstrating operational competence, we establish amateur radio as a serious participant in space networking.

---

## How Does This Compare to Existing Amateur Satellites?

With conventional linear transponders, the satellite is essentially a mirror — receiving and retransmitting in real time. Both stations must be active simultaneously. There's no onboard intelligence.

A DTN node is fundamentally different. It receives data, understands its structure, stores it in memory, and actively decides when to forward it. Sender and receiver need not be active simultaneously. The spacecraft becomes an intelligent network participant rather than a passive reflector.

This is the difference between a telephone (both parties present) and a post office (the letter waits for collection). For deep-space communications where simultaneous connectivity is often impossible, the post office model is essential.

---

## Partners and Future Development

RADIANT welcomes collaboration:

- **AMSAT-UK** provides organisational support and access to the amateur satellite community
- **AMSAT-DL** brings expertise in amateur satellite engineering  
- **Goonhilly Earth Station** provides professional ground station expertise

The project seeks partnerships with universities, CubeSat teams, microwave experimenters, and software developers interested in space networking.

Beyond individual DTN nodes, RADIANT's architecture is evolving toward Contact Plan as a Service — treating scheduled communication opportunities as shared network resources managed independently of any single DTN engine.

---

## How to Get Involved

**Packet radio operators:** Phase 1 uses standard 9600-baud equipment.
**QO-100 operators:** Phase 2 uses existing ground stations.
**Software developers:** Open-source codebase in Go, C++, and Rust.
**Microwave/EME operators:** Cislunar phases use similar techniques and equipment.

All software is available under MIT licence at https://radiant.amsat-uk.org

---

## Why This Matters

Building a delay-tolerant network across space combines radio engineering, software development, orbital mechanics, and protocol design — the multidisciplinary challenges that have always attracted the best in amateur radio.

As commercial space activities expand, demand for communications infrastructure beyond Earth will grow. Amateur radio has always been strongest when pioneering technology that later becomes mainstream. RADIANT positions amateur radio at the forefront of space networking using protocols that will connect future human settlements.

We are at a remarkable moment in space communications history. For the first time, the protocols that will underpin networking beyond Earth are being standardised in the open, with publicly available specifications and open-source implementations.

Amateur radio operators have always excelled at experimenting on the frontier. The frontier now extends beyond Earth, and RADIANT is how we extend with it.

---

**Contact:**  
David Johnson, G4DPZ  
Email: dave@g4dpz.me.uk  
Website: https://radiant.amsat-uk.org

*David Johnson, G4DPZ, is the project lead for RADIANT and Honorary Secretary of AMSAT-UK.*