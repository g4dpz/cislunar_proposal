# RADIANT: Building an Amateur Radio Network for the Solar System

**David Johnson, G4DPZ**

*An open-source project building the amateur radio network that works when the internet can't — from your back garden to the Moon*

---

## A Familiar Problem in an Unfamiliar Place

Imagine, a few years from now, you are an amateur radio operator. You have just finished setting up a ground station in your garden, pointed at the Moon. You type a short message into your terminal, press send, and watch as your data leaves Earth, travelling at the speed of light toward an amateur radio payload somewhere near the Moon. About 1.3 seconds later, it arrives. Another 1.3 seconds after that, an acknowledgement comes back. Your message has been stored onboard the spacecraft, waiting for the next contact window with a ground station on the far side of Europe to deliver it.

This is not science fiction. This is the goal of RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — and we are building it today.

RADIANT is an open-source project, supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station, that aims to bring space-grade networking technology to amateur radio. The project implements the same protocols that NASA and ESA are deploying for deep-space communications, adapted for amateur operators using equipment ranging from a Raspberry Pi and a Yaesu FT-817 to a dish antenna capable of receiving signals from the Moon.

**[IMAGE 1: Diagram showing the RADIANT phased roadmap — Phase 1 terrestrial packet radio links, Phase 1.5 via QO-100, Phase 3 a LEO satellite, and Phase 4 a node near the Moon. Each phase is a direct store-and-forward link between ground station and spacecraft. Caption: "The RADIANT roadmap: progressive phases from terrestrial radio links to a node near the Moon, each validating direct store-and-forward communication."]**

---

## Why the Internet Does Not Work in Space

Most of us take the internet for granted. When you browse a website, your computer opens a connection to a server, they exchange handshakes, and data flows in both directions almost instantaneously. This works because the internet was designed for an environment where connections are fast, reliable, and continuous.

Space is none of those things.

Consider what happens when you try to use conventional internet protocols (TCP/IP) to communicate with the Moon. The TCP handshake alone — the "hello, are you there?" exchange that happens before any data moves — requires three messages to travel between Earth and the Moon. At 1.3 seconds each way, that is nearly four seconds just to establish a connection. And that is the Moon, our nearest neighbour.

Mars, at its closest approach, is about three light-minutes away. At average distance, it is twelve minutes. A TCP handshake to Mars could take over an hour. Clearly, this approach does not scale.

The problem is not bandwidth. Modern radio links can achieve reasonable data rates even over enormous distances. The problem is latency — the time it takes for signals to travel at the speed of light — and intermittency. A satellite in low Earth orbit is only visible to a ground station for perhaps ten minutes per pass. A cislunar spacecraft may have its antenna pointed away from Earth for hours at a time. The links we take for granted on the ground simply do not exist in space.

---

## Delay-Tolerant Networking: A Postal Service for Data

The solution is an approach called Delay-Tolerant Networking, or DTN. Rather than assuming a continuous connection between sender and receiver, DTN works more like a postal service. Data is packaged into "bundles" (think of them as parcels), which are stored at each node in the network and forwarded to the next node whenever a communication opportunity arises.

If the next hop is not available — perhaps the satellite has not yet risen above the horizon, or the link is currently occupied — the node simply holds onto the bundle until the path opens up. When the bundle finally reaches its destination, the receiving node sends an acknowledgement back through the chain.

This store-and-forward model will be very familiar to anyone who remembers packet radio bulletin boards in the 1980s and 1990s. Messages were stored at each BBS node and forwarded overnight to the next station when propagation or scheduled links allowed. DTN formalises this concept using modern protocols standardised by the Internet Engineering Task Force (IETF) and the Consultative Committee for Space Data Systems (CCSDS) — the same body that standardises protocols for all major space missions.

The core protocols are:

- **Bundle Protocol version 7 (BPv7)**, defined in RFC 9171 — the networking layer that creates, stores, and delivers bundles
- **Licklider Transmission Protocol (LTP)**, defined in RFC 5326 — the transport layer that provides reliable delivery with deferred acknowledgement, specifically designed for links with long round-trip times
- **Contact Graph Routing (CGR)** — a routing algorithm that computes paths through the network based on predicted contact windows, much like a railway timetable

These are not experimental curiosities. NASA has been developing DTN since the early 2000s. The International Space Station has run DTN experiments. The Korean Pathfinder Lunar Orbiter (KPLO), currently in orbit around the Moon, carries Bundle Protocol as a development test objective. ESA's Moonlight programme — which is building communications and navigation infrastructure around the Moon — uses these same protocols as its foundation.

**[IMAGE 2: A simple diagram comparing TCP/IP (continuous connection, fails with delay) versus DTN (store-and-forward, tolerates delay and disruption). Perhaps showing a TCP handshake timing diagram over a lunar link versus DTN's "send, store, forward when available" approach. Caption: "TCP requires continuous connectivity. DTN stores data and forwards it when the next link becomes available — essential for space communications."]**

---

## What is Cislunar Space?

Before going further, it is worth explaining a term that appears frequently in this article. "Cislunar" refers to the region of space between Earth and the Moon — everything from low Earth orbit outward to the lunar vicinity. It is the space where most human activity beyond LEO will take place in the coming decades: the Artemis programme, the Lunar Gateway space station, commercial lunar landers, and eventually permanent settlements.

The cislunar region presents the communication challenges that DTN was designed to solve: delays of 1.3 seconds each way, orbital motion causing intermittent visibility, and limited power and antenna gain on spacecraft. It is also the region where amateur radio operators have a realistic opportunity to participate, given the right technology.

---

## How RADIANT Sends Data Over Amateur Radio

RADIANT uses the same KISS framing that TNCs have used for decades — so the radio side of things will be familiar to anyone who has operated packet. The difference is what sits inside those frames.

Instead of AX.25 packets, RADIANT carries DTN data directly in KISS frames. At the top level, messages are packaged as "bundles" using the Bundle Protocol (the space networking standard). Below that, LTP (Licklider Transmission Protocol) handles reliable delivery — designed specifically for links where acknowledgements take a long time to come back. And at the bottom, KISS feeds the data to your TNC and radio as normal.

The key advantage of this approach is simplicity. By skipping the AX.25 layer entirely (which would add overhead without benefit for DTN traffic), we get a cleaner, more efficient link. The system runs on NASA Glenn Research Centre's open-source HDTN software, though the architecture supports other DTN engines as well — giving operators flexibility in how they set up their stations.

---

## Station Identification: Callsigns in the Protocol

One of the key challenges in adapting space networking protocols for amateur radio is regulatory compliance — specifically, the requirement for station identification. Every amateur transmission must carry the operator's callsign.

RADIANT solves this elegantly by embedding callsigns directly into DTN Endpoint Identifiers (EIDs). Every bundle transmitted carries a source address like `dtn://g4dpz/service`, meaning the operator's callsign appears in every single piece of data sent over the air. This is not an afterthought bolted on to satisfy regulations — it is woven into the protocol's addressing scheme.

Additionally, the system transmits a plaintext beacon every ten minutes containing the station callsign, grid locator, and node type. Any station that can demodulate the signal can identify the transmitter, even if they cannot decode the DTN data itself.

This approach follows the precedent set by FT8, APRS, D-STAR, and Winlink — all of which embed callsigns within their protocol structures. All protocols and data formats used by RADIANT are publicly documented (via IETF RFCs and open-source code), satisfying the regulatory requirement that amateur radio protocols must be available for public inspection.

There is no encryption anywhere in the system. All data travels in the clear. Error correction coding is used where necessary to protect against corruption, but this does not obscure the meaning of communications. This keeps RADIANT fully compliant with ITU Radio Regulations and national amateur radio rules.

**[IMAGE 3: A simple diagram showing how callsigns are embedded in the DTN addressing scheme, with an example address like "dtn://g4dpz/service". Caption: "Station identification is built into the addressing scheme. Every transmission carries the operator's callsign as part of its source address."]**

---

## Contact Graph Routing: Scheduling by the Stars

In terrestrial networking, routers discover paths dynamically — your home router does not know in advance which route your packets will take. In space, the situation is quite different. We often know precisely when communication links will be available. We know when a satellite will rise above the horizon. We know when one spacecraft will have line-of-sight to another.

Contact Graph Routing exploits this predictability. A "contact plan" describes all known future communication opportunities: which nodes can talk to which other nodes, when the link starts, when it ends, what data rate is available, and what the propagation delay will be. The routing algorithm then computes optimal paths through this time-varying network, much like planning a journey using a train timetable.

For RADIANT's terrestrial phase, contact plans are relatively simple — ground stations may have always-on connectivity or scheduled operating hours. For the orbital phases, contact plans are computed from orbital predictions, giving precise pass windows for each ground station.

---

## The Five-Phase Roadmap

RADIANT follows a phased approach, with each phase validating critical technology needed for the next.

### Phase 1: Terrestrial Validation (In Progress)

The first phase uses equipment already familiar to many packet radio operators: a Raspberry Pi, a Mobilinkd TNC4, and a Yaesu FT-817 operating at 9600 baud. This validates the complete software stack over real amateur radio links before we commit to more complex (and expensive) hardware.

Phase 1 proves that DTN works over amateur radio. It establishes the baseline.

### Phase 1.5: QO-100 (Planned)

The Es'hail-2 geostationary satellite (QO-100) provides an ideal testbed for space-based DTN. With its always-on transponder and approximately 250 milliseconds one-way delay, it offers a genuine space environment without the complexity of orbital tracking. This phase validates LTP's deferred acknowledgement mechanism over a real space link — the critical stepping stone between ground-based millisecond delays and cislunar 1.3-second delays.

### Phase 2: CubeSat Engineering Model (Planned)

A ground-based "flatsat" — a bench-top replica of the flight hardware — using the same microcontroller and software that will eventually fly in orbit. This phase validates the software under realistic constraints: limited memory, simulated orbital passes, power cycling, and deliberate fault injection to prove the system recovers gracefully.

### Phase 3: LEO CubeSat (Planned)

The first RADIANT node in orbit. A CubeSat payload operating on 437 MHz UHF at 9.6 kbps, accessible to amateur ground stations worldwide with modest equipment (a Yagi antenna and SDR or TNC). This demonstrates ground-to-space DTN ping and store-and-forward messaging from orbit — a genuine first for amateur radio.

### Phase 4: Cislunar (Planned)

The ultimate goal: an amateur DTN node in the region between Earth and the Moon. Operating on S-band at 500 bits per second with strong forward error correction. This would be the first amateur-operated interplanetary-style communication system, proving that the amateur community can participate in the emerging architecture of space networking.

**[IMAGE 4: A timeline/roadmap diagram showing all five phases with their key milestones. Phase 1 (packet radio, 9600 baud), Phase 1.5 (QO-100), Phase 2 (engineering model), Phase 3 (LEO CubeSat, 437 MHz), Phase 4 (Cislunar, S-band). Caption: "The RADIANT roadmap: each phase validates technology needed for the next, building toward cislunar communications."]**

---

## What We Have Already Achieved

RADIANT is not merely a paper exercise. The project has a functioning three-node simulation that demonstrates store-and-forward with realistic propagation delays — 1.3 seconds for Earth-Moon distances, and configurable up to 12 minutes for Mars scenarios. We can observe the system managing deferred acknowledgements, computing routing paths, and delivering data successfully despite the delays.

The entire codebase is open-source under the MIT licence, with automated testing ensuring reliability with every change.

---

## The Bigger Picture: A Solar System Internet

RADIANT is not happening in isolation. In September 2023, the Internet Society's Interplanetary Networking Special Interest Group (IPNSIG) — co-chaired by Vint Cerf, one of the fathers of the internet — published a report on "Solar System Internet Architecture and Governance." Their recommendation: the same protocols RADIANT implements should form the basis for all communications beyond Earth.

The report envisions a "network of networks" in space, much like the terrestrial internet — multiple organisations operating their own networks, interconnecting through standard protocols and peering agreements. It calls for multistakeholder governance, meaning that decisions about space networking architecture should not be made solely by space agencies. Other stakeholders — including the amateur radio community — have a legitimate role to play.

This is not unprecedented. Amateur radio operators were experimenting with packet networking before the commercial internet existed. AMSAT launched the first amateur satellite (OSCAR 1) in 1961, decades before commercial satellite constellations. The amateur community has a long tradition of pioneering technology that later becomes mainstream.

RADIANT is how we earn our place in the emerging Solar System Internet. By implementing the same protocols, demonstrating operational competence, and contributing open-source tools and operational data, we establish amateur radio as a serious participant in space networking — not merely an observer.

---

## How Does This Compare to What We Already Do?

Some readers may wonder how DTN differs from existing amateur satellite operations. The distinction is fundamental.

With a conventional linear transponder (like those on QO-100 or most LEO amateur satellites), the satellite is essentially a mirror. It receives a signal and retransmits it in real time. Both the sending and receiving stations must be active simultaneously. There is no intelligence onboard — the satellite does not understand or store the data passing through it.

A DTN node is fundamentally different. It receives data, understands its structure, stores it in onboard memory, and actively decides when and how to forward it. The sender and receiver need not be active at the same time. The spacecraft becomes an intelligent participant in the network rather than a passive reflector.

This is the difference between a telephone (both parties must be present) and a post office (the letter waits until the recipient collects it). For deep-space communications where simultaneous connectivity is often impossible, the post office model is essential.

---

## Ground Station Requirements

One of RADIANT's design goals is accessibility. Participation does not require exotic equipment:

**Phase 1 (Terrestrial):** Any 9600-baud packet radio setup. A Raspberry Pi, Mobilinkd TNC4, and a VHF/UHF radio with a data port is sufficient. Many RSGB members will already have suitable equipment.

**Phase 1.5 (QO-100):** A standard QO-100 narrowband ground station — typically a 60-90cm offset dish, an SDR or dedicated transceiver, and appropriate feeds for 2.4 GHz uplink and 10.45 GHz downlink.

**Phase 3 (LEO):** A UHF Yagi antenna (the same equipment used for existing amateur satellite work on 435 MHz) with a TNC or SDR receiver. The LEO CubeSat's link budget provides over 30 dB of margin, meaning modest ground stations can participate comfortably.

**Phase 4 (Cislunar):** This is where larger equipment becomes necessary — a dish antenna with low-noise amplification, operating on S-band. This is comparable to existing EME (moonbounce) station capabilities, and could also involve collaboration with university or institutional ground stations.

The software is entirely open-source and runs on standard computing hardware. Ground station operators will be able to download and install the RADIANT client software, configure their callsign and station parameters, and begin participating in the network.

**[IMAGE 5: A photograph or diagram of a Phase 1 ground station setup — Raspberry Pi, TNC4, and radio — showing how accessible the entry point is. Alternatively, a montage showing the range of ground station tiers from a simple Yagi setup to a larger dish. Caption: "Entry-level RADIANT participation requires only a Raspberry Pi, a TNC, and a 9600-baud radio — equipment many operators already own."]**

---

## Regulatory and Compliance Considerations

RADIANT has been designed from the ground up to operate within amateur radio regulations:

- **No encryption** — all data transmitted over amateur links is unencrypted, fully compliant with ITU Radio Regulations Article 25 and national rules. The sole exception, as permitted by regulations, would be encrypted telecommand for satellite command and control.
- **Published protocols** — all protocols are publicly documented through IETF RFCs and open-source code. Any operator or regulator can inspect exactly how data is encoded and transmitted.
- **Station identification** — callsigns are embedded in every transmission through the EID scheme, plus periodic plaintext beacons, exceeding the minimum regulatory identification requirements.

---

## Partners and Collaboration

RADIANT is a collaborative effort:

- **AMSAT-UK** provides organisational support, access to the amateur satellite community, and a platform for engaging UK operators
- **AMSAT-DL** brings deep expertise in amateur satellite engineering (they built the amateur transponders on QO-100) and connections to the European amateur satellite community
- **Goonhilly Earth Station** provides professional ground station expertise and potential access to large-aperture antennas for cislunar phases

The project is also seeking collaboration with universities (for ground station time and research partnerships), CubeSat teams (for potential hosted payload opportunities), microwave experimenters, and anyone with an interest in space networking, distributed systems, or protocol engineering.

---

## How to Get Involved

RADIANT welcomes participation at every level:

- **Packet radio operators**: Phase 1 testing uses standard 9600-baud equipment. If you have a packet station, you have the hardware.
- **QO-100 operators**: Phase 1.5 will use existing QO-100 ground station setups.
- **Software developers**: The codebase is open-source (Go, C++, Rust). Contributions are welcome — from protocol implementation to ground station software to web interfaces.
- **Microwave and EME operators**: The operational patterns and ground station requirements for cislunar DTN closely mirror moonbounce practices. Your experience is directly relevant.
- **University groups and clubs**: RADIANT provides a practical framework for teaching space communications, networking protocols, and systems engineering.

All software is freely available under the MIT licence. Documentation, specifications, and design documents are public. The project website at https://radiant.amsat-uk.org provides current status, documentation, and contact details.

---

## Link Budgets: Can Amateurs Really Do This?

The link budgets close comfortably for the earlier phases. The LEO CubeSat at 437 MHz provides approximately 31 dB of margin with just 2 watts transmit power and a ground-based Yagi — well within reach of any station currently tracking amateur satellites.

The cislunar phase is tighter: approximately 7 dB margin at 500 bits per second using S-band with strong forward error correction. This is achievable with larger ground station dishes, comparable to what EME operators already use. The key insight is that data rate trades against distance — 500 bps is slow, but DTN's store-and-forward model means messages accumulate across multiple contact windows.

---

## A Future GEO Backbone

Looking beyond the immediate roadmap, RADIANT has proposed a DTN payload for the next-generation GEO amateur satellite (the "Future GEO" project). A DTN payload onboard a geostationary satellite would transform it from a real-time transponder into an intelligent store-and-forward node — always available, delivering bundles to ground stations whenever they come online. Unlike a conventional transponder, the sender and receiver need not be active simultaneously.

---

## Why This Matters

Building a delay-tolerant network across space combines radio engineering, software development, orbital mechanics, and protocol design into one integrated challenge — the kind of multidisciplinary problem that has always attracted the best in amateur radio.

More broadly, as commercial space activities expand, the demand for communications infrastructure beyond Earth will grow. Amateur radio has always been strongest when pioneering technology that later becomes mainstream. The IPNSIG Solar System Internet report explicitly calls for multistakeholder governance of space networking. Amateur radio operators are a legitimate stakeholder — but only if we demonstrate operational competence. RADIANT is how we build that credibility.

---

## Looking Forward

We are at a remarkable moment in the history of space communications. For the first time, the protocols that will underpin networking beyond Earth are being standardised, implemented, and deployed — not behind closed doors in space agencies, but in the open, with publicly available specifications and open-source reference implementations.

Amateur radio operators have always been at their best when experimenting at the frontier — from spark-gap transmitters to packet radio to software-defined transceivers. The frontier is now extending beyond Earth, and RADIANT is how we extend with it.

The technical foundations are proven. The protocols are standardised. The software exists. What remains is for the amateur community to do what it has always done: build, test, improve, and share.

If you would like to be part of building amateur radio's first interplanetary network, we would be glad to hear from you.

---

**Contact:**  
David Johnson, G4DPZ  
Email: dave@g4dpz.me.uk  
Website: https://radiant.amsat-uk.org  
Source code: https://github.com/g4dpz/cislunar_proposal

---

*David Johnson, G4DPZ, is the project lead for RADIANT. He is Honorary Secretary of AMSAT-UK and a Senior Software Engineer at Goonhilly Earth Station, where he works on commercial lunar communications.*
