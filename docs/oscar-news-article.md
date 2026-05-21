---
title: "RADIANT: An Amateur Radio Pathway to Cislunar Delay-Tolerant Networking"
author: "David Johnson, G4DPZ"
date: "2026"
geometry: margin=2.5cm
fontsize: 11pt
---

For more than two decades, amateur satellites have demonstrated how radio amateurs can contribute meaningfully to space communications research. RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — aims to take that tradition one step further: building the first amateur-operated Delay-Tolerant Networking (DTN) system capable of extending into cislunar space.

The long-term ambition is bold but technically grounded. RADIANT's roadmap culminates in an amateur DTN node operating beyond Earth orbit, using protocols derived from the same networking standards employed by NASA and ESA for deep-space communications. The CubeSat and cislunar missions are the destination; the earlier terrestrial and GEO phases exist to validate each technical building block before flight.

The project is built on NASA Glenn Research Center's High-rate Delay Tolerant Networking (HDTN) software stack, implementing Bundle Protocol v7 (RFC 9171) and the Licklider Transmission Protocol (RFC 5326). RADIANT is fully open-source under the MIT licence and is supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station.

## Why DTN Matters in Space

Conventional TCP/IP networking assumes that a continuous path exists between sender and receiver. That works well on the terrestrial Internet but breaks down rapidly in space environments where links are intermittent, propagation delays are large, and connectivity may disappear entirely for extended periods.

Delay-Tolerant Networking solves this by using a store-and-forward model. Data is encapsulated into bundles, stored at each node, and forwarded only when the next scheduled contact becomes available. In many ways this resembles classic packet radio BBS forwarding, but formalised into modern IETF and CCSDS standards suitable for spaceflight operations.

The need becomes obvious when examining real orbital scenarios. A LEO satellite may only be visible for five to ten minutes per pass. GEO systems introduce approximately 500 ms round-trip delay, while Earth–Moon communications incur roughly 1.3 seconds one-way propagation time. Traditional networking protocols cannot function reliably under such conditions.

RADIANT uses the Licklider Transmission Protocol (LTP) beneath BPv7. LTP was designed specifically for long-delay environments, using deferred acknowledgements and checkpoint-based reliability mechanisms that tolerate very high latency. Routing is handled using Contact Graph Routing (CGR), where transmissions are scheduled according to predicted contact windows derived from orbital data, antenna availability, and mission timelines.

## The CubeSat Mission — DTN in Low Earth Orbit

The first orbital target for RADIANT is a CubeSat-class DTN payload in Low Earth Orbit. The primary mission objective is straightforward but significant: demonstrate reliable ground-to-space DTN operation using amateur radio infrastructure.

The planned payload uses an STM32U585 ultra-low-power ARM Cortex-M33 microcontroller operating at 160 MHz, with 2 MB of flash memory and 786 KB of SRAM. The device includes hardware cryptographic acceleration and ARM TrustZone support while maintaining extremely low power consumption.

The RF subsystem centres on a flight-qualified IQ transceiver IC interfacing directly with the STM32 via DAC and ADC paths. Initial operations are planned on the 437 MHz amateur satellite band at 9.6 kbps using GMSK or BPSK modulation — intentionally chosen to remain accessible to existing amateur ground stations worldwide.

Link budget analysis indicates substantial operational margin. With a 2 W transmitter, omnidirectional spacecraft antenna, and a modest 12 dBi ground Yagi, the predicted margin exceeds 30 dB under nominal conditions. This allows experimentation with robust DTN operation without requiring exotic ground hardware.

Operationally, the spacecraft acts as a delay-tolerant store-and-forward node. Messages uploaded during one pass may be stored in persistent non-volatile memory and delivered during later passes. DTN ping operations validate end-to-end protocol behaviour, while CGR uses TLE/SGP4 orbital predictions to schedule contacts automatically.

Persistent bundle storage ranges from 64 to 256 MB of external non-volatile memory, sufficient to survive power interruptions and prolonged communication gaps. Power consumption targets are modest — approximately 5–10 W average — aided by the STM32U585's Stop 2 low-power mode drawing around 16 µA between communication windows.

The wider vision includes a distributed network of amateur-operated ground stations running open-source client software. The mission is intended not only to prove the hardware but also to generate a public operational dataset for future amateur and academic DTN research.

## The Cislunar Mission — Amateur Networking Beyond Earth Orbit

Beyond LEO lies the project's most ambitious phase: a true amateur-operated cislunar DTN node.

The current concept involves a hosted payload aboard a cislunar spacecraft or lunar CubeSat, potentially operating in highly elliptical Earth orbit or during lunar transfer trajectories. Unlike the LEO mission, the cislunar system moves into a genuine deep-space communications regime.

The RF architecture transitions to S-band around 2.2 GHz using BPSK modulation with strong LDPC or Turbo forward error correction. Due to the enormous free-space losses involved, data rates are intentionally conservative — approximately 500 bps. Preliminary link analysis suggests feasibility using a 5 W transmitter, a 10 dBi spacecraft patch antenna, and 35 dBi ground stations using 3–5 metre dishes, yielding approximately 7 dB link margin.

At average lunar distance the one-way propagation delay is approximately 1.3 seconds. Such latency fundamentally changes how networking protocols behave and provides an ideal real-world test environment for DTN and LTP.

Ground infrastructure for this phase would involve Tier 3/4 stations employing larger dishes, phased arrays, and low-noise microwave front ends. Potential experiments include Earth–Moon DTN ping measurements, delay-tolerant file transfer, and resilience testing under prolonged outages and extreme latency.

The engineering challenges are substantial: radiation tolerance, tight power budgets, antenna pointing constraints, and international frequency coordination all become critical. RADIANT addresses these risks through conservative data rates, robust FEC schemes, and a phased validation approach that incrementally proves each subsystem before deployment.

Discussions are underway regarding potential ESA ARTES support and partnerships with university ground station networks.

## Building the Foundation — A Phased Development Strategy

The most important aspect of RADIANT is that the deep-space mission is not being approached as a single leap. Every earlier phase validates technologies directly relevant to eventual cislunar operation.

Phase 1, currently underway, uses Raspberry Pi systems with Mobilinkd TNC4 modems and Yaesu FT-817 transceivers operating at 9600 baud G3RUH packet on VHF/UHF. This is not merely a software exercise — it validates the complete protocol stack over real amateur radio links with all the impairments that entails: fading, interference, and genuine disruption.

The terrestrial nodes implement the full operational feature set planned for the space missions. Bundles are stored persistently on the local filesystem, surviving power cycles and process restarts. A contact plan manager maintains scheduled communication windows between nodes, and the system transmits queued bundles in strict priority order (critical, expedited, normal, bulk) during each window. LTP provides reliable transfer with deferred acknowledgment, automatically retransmitting unacknowledged segments in subsequent contact windows. Rate limiting protects the bundle store from flooding, and expired bundles are automatically evicted.

Current testing between G4DPZ and M0XER has demonstrated functional store-and-forward delivery and DTN ping operation. The system correctly handles link interruptions — if the TNC connection drops mid-contact, bundles are retained and retried during the next window. This behaviour directly mirrors what the CubeSat payload must do when a ground pass ends or a link degrades.

The planned Phase 1.5 introduces QO-100 as the first genuine space-based DTN experiment. Using the Es'hail-2 narrowband transponder, the project will validate DTN operation over a real GEO satellite path with authentic 500 ms round-trip delay. This phase acts as a crucial bridge between terrestrial tests and orbital hardware.

Phase 2 introduces a CubeSat engineering model using the STM32U585 flight computer alongside an Ettus B200mini SDR for representative IQ baseband testing. Simulated orbital passes, power budget profiling, thermal-vacuum preparation, and fault injection testing will validate the exact software and hardware stack intended for flight.

An important design philosophy throughout is that the terrestrial network intentionally mirrors the eventual cislunar communications chain:

> Mission Operations → Ground Gateway → Amateur RF Link → Relay Node → Payload Endpoint

Every demonstration therefore exercises the same DTN behaviour, contact scheduling logic, and store-and-forward mechanisms needed for deep-space operation.

The project has already achieved several notable technical milestones, including a functioning three-node cislunar simulation incorporating realistic packet-level propagation delays ranging from lunar-scale 1.3-second paths to Mars-scale multi-minute round-trip times. CGR is successfully computing multi-hop relay paths while LTP manages RTTs up to 24 minutes. Supporting infrastructure includes a custom C++17 KISS convergence layer adapter for HDTN, a Go-based orchestration system, property-based testing, and a full CI pipeline.

## Protocol Stack and Regulatory Compliance

RADIANT adopts a deliberately simplified protocol architecture:

> Application → BPv7 → LTP → KISS → G3RUH (9600 baud)

One notable design choice is the elimination of AX.25 entirely. LTP packets are encapsulated directly within KISS framing, reducing overhead by approximately 15 bytes per frame and improving throughput efficiency by roughly 10%.

Regulatory compliance remains central to the design. Station identification is achieved using callsign-based DTN Endpoint Identifiers such as `dtn://g4dpz-1`. Every transmitted bundle therefore contains explicit operator identification within the source EID. Additional plaintext beacon bundles transmitted every ten minutes include callsign, node name, EID, timestamp, and Maidenhead locator — ensuring that any station demodulating the signal can identify the transmitter, even when the wire format carries only opaque numeric `ipn://` routing addresses.

The system does not use encryption and remains fully compliant with amateur radio regulations. The project plans to publish a formal protocol definition document describing the callsign EID convention, SSID allocation, service demultiplexing, and beacon formats. This follows the established precedent set by APRS, FT8, D-STAR, and Winlink, all of which rely on publicly documented protocol specifications with callsign identification defined within them.

## Getting Involved

RADIANT is designed as a community project from the outset. Participation ranges from modest terrestrial stations using a Raspberry Pi and 9600 baud packet equipment, through to microwave and deep-space capable stations employing larger dishes and phased arrays.

The project is actively seeking collaboration from amateur radio clubs, universities, CubeSat teams, packet radio operators, microwave experimenters, EME and weak-signal operators, and researchers interested in space networking.

All software and documentation are open-source and publicly available.

**Website:** https://radiant.amsat-uk.org

**Contact:** dave@g4dpz.me.uk / G4DPZ

Whether contributing code, providing ground station support, assisting with protocol development, or collaborating on future flight hardware, new participants are very welcome.
