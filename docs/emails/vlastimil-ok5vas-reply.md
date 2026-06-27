# Reply to Vlastimil OK5VAS

**To:** Vlastimil (OK5VAS)  
**From:** Dave Johnson (G4DPZ)  
**Subject:** Re: PacketRF, lwBP, and RADIANT — lots of overlap here

---

Hi Vlastimil,

No apology needed for the timing — I understand completely. This is a hobby project for most of us too, and a thoughtful reply is always better than a quick one. Thanks for taking the time to look through RADIANT properly.

I've had a look at PacketRF and the uart.cz article — impressive work getting NPR interoperability on RP2350. And lwBP sounds exactly like the kind of constrained-node BPv7 implementation the amateur DTN ecosystem needs.

Let me respond to your points in order:

**On heterogeneous link layers:**

This is precisely the architecture we're building. Our DTN abstraction layer (just completed in Rust) provides a vendor-neutral management plane that's designed to support multiple DTN engines and multiple convergence layers simultaneously. The canonical config model defines nodes and neighbors with typed CL links — currently LTP/UDP, TCPCLv4, KISS (for TNC hardware), and plain UDP. Adding NPR as another CL type would be a natural extension.

The key insight we're working from is: BPv7/LTP is the store-and-forward layer, and the convergence layer below it should be pluggable. Your PacketRF links (NPR, future AX.25, future LoRa) would each be a convergence layer adapter in our model. The DTN stack doesn't care how the bytes get from A to B — it handles the disruption tolerance, priority, storage, and routing above whatever link layer is available.

**On routing and addressing:**

Your observation about CGR vs amateur networks is spot on. CGR works when you have a predictable contact schedule (satellite passes, planned links). For informal amateur networks where links come and go unpredictably, you need something different.

We're currently using CGR for the orbital phases (where contacts are schedulable) and static routing for terrestrial (where links are always-on or manually managed). But the gap you've identified — opportunistic routing for ad-hoc amateur networks — is real. Some options I've been thinking about:

1. **Epidemic/spray-and-wait** for low-connectivity networks (works but wasteful)
2. **PRoPHET-style** probabilistic routing based on encounter history
3. **Hybrid**: CGR for planned contacts (satellite passes, scheduled QSOs), opportunistic for the rest

Your idea of structured IPN addressing (like IP subnetting for amateur DTN) is interesting. The callsign-based EID scheme we use (`dtn://callsign-ssid/service`) already provides global uniqueness, but it doesn't give you topological routing hints. A hierarchical IPN space (region/country/network/node) could enable prefix-based routing decisions without full global knowledge. Worth exploring.

**On LTP-over-KISS vs AX.25:**

We've dropped AX.25 entirely. Our protocol stack is BPv7 → LTP → KISS framing → radio. LTP segments are wrapped directly in KISS frames with no AX.25 header. Station identification is via callsign-embedded DTN EIDs (`dtn://g4dpz-1/service`) in every bundle, plus periodic beacon bundles every 10 minutes.

The architecture document explaining this is at `docs/LTP-KISS-ARCHITECTURE.md` in the repo. For your 1200 baud AFSK work, the same approach would work — LTP segments in KISS frames over Bell 202. The framing overhead is just 3 bytes (FEND + CMD + FEND) versus 18+ bytes for AX.25 headers, which matters a lot at 1200 baud.

For G3RUH 9600 baud on RP2350 — that's definitely feasible. The G3RUH GFSK demodulation is well-documented and there are open implementations. At 9600 baud you have enough processing headroom on RP2350 for the modem plus a lightweight BPv7 stack.

**On lwBP:**

Very interested to learn more when you publish it. We're targeting Hardy (Rust, `no_std`) and µD3TN (C, lightweight) as our constrained-platform engines. A C++ implementation for RP2350-class hardware fills a slightly different niche — even more constrained than our STM32U585 target (which has 786 KB SRAM). If lwBP can interoperate with ION and Hardy over standard LTP/UDP, that's a third implementation for our interoperability testing, which would be extremely valuable.

We already have ION-DTN 4.1.4 and Hardy 0.1.0 starting successfully from our abstraction layer's generated configs, and we're working on cross-engine bundle exchange tests. Adding lwBP as a third engine would strengthen the interop story significantly.

**Concrete collaboration opportunities:**

1. **NPR as a convergence layer** — once PacketRF is stable, adding NPR as a CLA type in our abstraction layer model
2. **lwBP interoperability** — testing bundle exchange between lwBP (RP2350), ION-DTN (Linux), and Hardy (Linux/embedded Rust)
3. **Routing research** — jointly exploring opportunistic/hybrid routing for amateur DTN beyond CGR
4. **1200 baud DTN** — your Bell 202 work + our LTP-over-KISS protocol = DTN at 1200 baud on legacy VHF packet hardware
5. **Addressing conventions** — collaborating on a structured IPN address space proposal for amateur DTN

I'd definitely like to have a call to discuss further. We also have a growing group of collaborators: Leigh Torgerson (ex-JPL, ION-DTN testbed lead), Jorge Amodio (IPNSIG), Loïc F4JXQ (Hardy testing, FOSM-1 satellite), plus AMSAT-UK/AMSAT-DL. You'd be very welcome in that group.

Looking forward to seeing lwBP published.

73,
Dave, G4DPZ
