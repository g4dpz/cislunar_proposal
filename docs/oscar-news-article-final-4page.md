# RADIANT: Amateur Radio's Cislunar Network

**David Johnson, G4DPZ**

RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — demonstrates amateur radio's capability in space-grade networking protocols, which has two long term mission ambitions: 
 - DTN payload on LEO CubeSat
 - ultimately cislunar networking node.

This project is a collaboration between AMSAT-UK, AMSAT-DL and Goonhilly Earth Station and implements Bundle Protocol v7 and Licklider Transmission Protocol — the same standards NASA and ESA use for deep-space missions.

## Space Internet Standards

The protocols RADIANT uses aren't experimental — they're becoming the standard for space communications. In 2023, the Internet Society's Interplanetary Chapter published a Solar System Internet governance framework co-authored by Vint Cerf, widely recognized as one of the "fathers of the Internet" for creating TCP/IP. Their recommendation: DTN and Bundle Protocol should be the foundation for all space communications that exceed IP's capabilities.

Both ESA's Moonlight programme and NASA's LunaNet specification mandate these same protocols. Through RADIANT, amateur radio participates as a stakeholder in this emerging interplanetary infrastructure.

## Technical Approach

Internet protocols assume continuous connectivity. TCP requires multiple round trips — acceptable at millisecond delays but unworkable when satellites are visible briefly or the Moon is 1.3 light-seconds away.

DTN uses store-and-forward like 1980s packet BBS. Bundles persist at nodes until links become available. LTP provides reliable delivery with deferred acknowledgments. Contact Graph Routing schedules transmissions predictably.

These aren't experimental — ISS runs DTN today. Korea's lunar orbiter tested Bundle Protocol at lunar distance in 2024.

RADIANT carries DTN directly in KISS frames, eliminating AX.25 entirely — 15 bytes saved per frame (~10% improvement). Station identification embeds in DTN Endpoint Identifiers (`dtn://g4dpz-1/service`). Amateur callsigns solve global identifier allocation — no new registry required.

**[FIGURE 1: RADIANT Phase 1 ground station - Raspberry Pi running DTN software with Yaesu transceivers and TNC interfaces, validating space-grade networking protocols over standard 9600-baud packet radio links]**

## Four-Phase Validation

**Phase 1 — Terrestrial (Complete):** RPi, TNC4, FT-817 at 9600 baud validated complete software stack over real amateur radio links.

**Phase 2 — QO-100 (In Progress):** DTN through Es'hail-2 narrowband transponder at 9600 bps provides space environment validation.

**Phase 3 — CubeSat Engineering:** Ground flatsat using flight STM32U585 with thermal/radiation testing.

**Phase 4 — Missions:** LEO CubeSat followed by cislunar operations.

## LEO CubeSat Mission

Primary goal: amateur DTN payload demonstrating ground-to-space messaging and orbital store-and-forward.

- STM32U585 microcontroller (160 MHz, 2 MB flash)
- 437 MHz UHF at 9.6 kbps GMSK
- 31 dB link budget margin with standard UHF Yagi
- 64-256 MB persistent storage
- Compatible with existing amateur satellite equipment

Ground operations leverage distributed amateur satellite community through shared contact plans.

## Cislunar Mission

Phase 4 extends DTN to cislunar space where Artemis and commercial lunar missions operate.

- S-band 2.2 GHz at 500 bps, 5W + 10 dBi spacecraft antenna
- 35 dBi ground dish (3-5m, EME-class), 7 dB link margin
- 2.6-second lunar round-trip delay handling

Success positions amateur radio as permanent interplanetary communications stakeholder.

## Current Status

RADIANT operates three-node cislunar simulation with true 1.3-second Earth-Moon delays. Contact Graph Routing computes multi-hop paths while LTP manages extended round-trips. The project uses ION-DTN as the primary implementation, validating compatibility with the NASA standard DTN stack.

Project gained visibility through EMF Camp 2026 "Internet for the Solar System" presentation. All code open-source under MIT license.

## Building Networks

Long-term vision: robust DTN networks both terrestrial and cislunar. Terrestrial provides foundation while space-based relays extend coverage and emergency backup.

We're seeking amateur satellite operators and ground station collaborators internationally. University partnerships offer students space-grade networking experience while validating amateur radio's technical contributions.

Live demonstrations provide STEM outreach, inspiring space careers by showing amateur radio's contribution to professional standards.

## Get Involved

**Phase 1:** 9600-baud packet stations  
**Phase 2:** Standard QO-100 setups  
**Phase 3/4:** UHF satellite operators, EME stations

Software developers (Rust/Go/C++), microwave operators, amateur clubs, universities, and CubeSat teams welcome.

Amateur radio pioneered packet radio and satellites before commercial adoption. RADIANT positions amateur radio at the interplanetary communications forefront using protocols that will connect future human settlements.

**Contact:** David Johnson, G4DPZ • dave@g4dpz.me.uk • https://radiant.amsat-uk.org