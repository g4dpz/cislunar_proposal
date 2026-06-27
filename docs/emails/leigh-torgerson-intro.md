# Draft Email: Leigh Torgerson Introduction

**To:** Leigh Torgerson  
**From:** Dave Johnson (G4DPZ)  
**Subject:** RADIANT — amateur radio DTN testbed + DTN implementation interoperability

---

Hi Leigh,

Thanks for connecting — I appreciate Jorge making the introduction.

I'm leading a project called RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed) which is bringing BPv7/LTP to amateur radio, phased from terrestrial ground stations through LEO CubeSat to eventually a cislunar node. We're supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station, and we've recently been introduced to IPNSIG through Jorge and Vint.

The short version: we have a working 3-node DTN simulation with store-and-forward relay and CGR, currently using NASA Glenn's HDTN as our DTN engine. Phase 1 terrestrial testing is in progress using LTP wrapped directly in KISS framing over 9600 baud amateur radio links, with callsign-embedded DTN EIDs for station identification.

One area I'd particularly like to explore is **DTN implementation interoperability**. RADIANT's architecture is deliberately engine-agnostic — we've built an abstraction layer that's designed to support multiple DTN implementations behind a common interface. The implementations we're looking at are:

- **ION-DTN** (JPL) — the reference implementation, extensive flight heritage
- **HDTN** (NASA Glenn) — our current engine, high-rate C++17 implementation
- **µD3TN** (D3TN GmbH) — lightweight, space-tested, candidate for constrained flight hardware
- **ESA-DTN** — ESA's implementation for European missions
- **Hardy** — modular Rust BPv7 implementation with `no_std` core, candidate for microcontroller flight software

The amateur radio environment is interesting for interop testing because:

1. We have real RF links with real propagation characteristics, not just loopback
2. Different ground stations could run different engines (some operators will prefer ION, others HDTN or µD3TN)
3. A flight payload might use a different engine than ground infrastructure (Hardy or µD3TN on an STM32, HDTN on ground stations)
4. The constrained bandwidth (9600 baud to 500 bps) exercises edge cases in LTP segmentation and session management that high-rate links don't expose
5. BPv7 and LTP are wire-format standards — if implementations are conformant, they should interoperate regardless of which engine generated the bundle

We'd like to develop a concrete interop test plan: can an ION node exchange bundles with an HDTN node over LTP? Does µD3TN's LTP implementation handle the same edge cases as ION's? Do all five implementations agree on EID resolution, CGR contact plan interpretation, and LTP session behaviour under the same timing constraints?

Given your experience with ION-DTN at JPL, I'd value your perspective on:

- Known interop challenges between ION and other implementations
- Whether there are existing interop test suites or conformance tests we should be aware of
- Any gotchas in LTP session management that differ between implementations
- Whether JPL would have interest in an amateur radio network providing independent interop validation data

Happy to share more detail on the architecture, or jump on a call if that's easier. Our project site is https://radiant.amsat-uk.org.

73,
Dave, G4DPZ
