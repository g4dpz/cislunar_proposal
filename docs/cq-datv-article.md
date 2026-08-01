# DTN on QO-100: Store-and-Forward Data via the Wideband Transponder

**David Johnson, G4DPZ**

*How amateur delay-tolerant networking could ride alongside DATV through Es'hail-2 — and what you need to try it*

---

If you operate DATV through QO-100's wideband transponder, you already have most of what you need to participate in something quite different: a delay-tolerant data network that uses the same transponder bandwidth to move files, messages, and telemetry across continents — without requiring both stations to be online at the same time.

RADIANT — the Radio Amateur Delay-tolerant Interplanetary Networking Testbed — is an open-source project building amateur radio's first space networking infrastructure. One of its phases uses QO-100 as a testbed for protocols designed for deep-space communications: Bundle Protocol version 7 and Licklider Transmission Protocol, the same standards NASA uses on the International Space Station and ESA plans for its Moonlight lunar programme.

The idea is straightforward. You transmit a data bundle — a file, a message, a chunk of telemetry — up to Es'hail-2. On the other end, a ground station receives it, stores it, and delivers it when the intended recipient comes online. If nobody is listening right now, the bundle waits. It is a post office in geostationary orbit.

**[FIGURE 1: Simple diagram showing two ground stations, Es'hail-2 in the middle, with a bundle being transmitted up from one station, through the transponder, and down to the other. A clock icon indicates store-and-forward timing. Caption: "DTN store-and-forward via QO-100: the bundle waits at the receiving node until the destination is ready."]**

---

## Why QO-100 for DTN?

The geostationary link to Es'hail-2 introduces approximately 250 milliseconds of one-way delay (500 ms round-trip). That is enough to exercise the delay-tolerant aspects of the protocol stack authentically. LTP — the transport layer beneath Bundle Protocol — was specifically designed for links where acknowledgements take a noticeable time to return. QO-100 provides a real space link without the complexity of orbital tracking or brief pass windows.

For DATV operators, the wideband transponder at 10.491 GHz downlink (2.409 GHz uplink) offers significant bandwidth. A narrow DVB-S2 carrier — say 333 kS/s, comparable to what many DATV operators already use — could carry DTN traffic at meaningful data rates. You are not limited to the narrowband transponder's audio bandwidth. The wideband side offers enough capacity for file transfers, image relay, and accumulated telemetry delivery.

The transponder is always there. Always visible from Europe. No tracking required. You point your dish at 25.9°E and the link is available 24 hours a day.

---

## What the Data Path Looks Like

Here is how a DTN bundle travels through QO-100:

1. **Your station** — a Raspberry Pi or laptop runs the RADIANT software, which manages a DTN engine (ION-DTN, Hardy, or µD3TN). You compose a bundle — perhaps a JPEG image, a text message, or a telemetry file — addressed to a destination node by its callsign.

2. **Uplink** — the bundle is segmented by LTP, wrapped in baseband frames, and transmitted on your 2.4 GHz uplink carrier to Es'hail-2. The transponder simply translates it to 10 GHz and retransmits. No processing onboard — it is a bent-pipe transponder like any other.

3. **Downlink** — a receiving ground station demodulates the 10 GHz signal and reassembles the LTP segments into complete bundles. If the destination node is local, delivery is immediate. If not, the bundle is stored and forwarded via terrestrial links (internet, packet radio, or another RF hop) to its final destination.

4. **Store-and-forward** — if the receiving station has a bundle for a node that is not currently connected, it holds it. When that node comes online — hours or days later — the stored bundle is delivered. No data is lost to timing.

The key difference from a live DATV QSO: both parties do not need to be active simultaneously. You can transmit bundles at 3 AM, and the recipient collects them at noon the next day.

---

## Equipment You Already Have

If you run DATV through QO-100's wideband transponder, your RF chain is ready:

**Uplink (2.4 GHz):**
- Offset dish (80 cm–1.2 m)
- DATV exciter or SDR (PlutoSDR, Adalm-Pluto, LimeSDR) generating a DVB-S2 carrier
- Power amplifier (a few watts is sufficient for narrow carriers)
- Appropriate feed for 2.4 GHz

**Downlink (10.491 GHz):**
- The same dish with a modified LNB or dedicated 10 GHz feed
- SDR receiver or dedicated DVB-S2 demodulator (MiniTiouner, TBS card, or software receiver)

**Computing:**
- Raspberry Pi 4/5 or any Linux machine
- RADIANT software (open-source, MIT licence)
- A DTN engine: ION-DTN for NASA compatibility, Hardy for a lightweight Rust implementation, or µD3TN for minimal resource usage

The difference from your DATV setup is purely in software. Instead of encoding video with OBS and transmitting an MPEG-TS stream, you run the RADIANT client, which manages bundle creation, LTP segmentation, and contact scheduling. The RF path is identical.

For narrowband QO-100 operators, participation is also possible at lower data rates through the narrowband transponder — useful for small messages and beacons, though throughput is naturally limited.

**[FIGURE 2: Block diagram of a typical ground station — SDR/PlutoSDR → PA → dish → Es'hail-2 → LNB → SDR receiver → Raspberry Pi running RADIANT. Caption: "A DATV-capable QO-100 station needs only the RADIANT software to become a DTN node."]**

---

## What Can You Send?

Bundles can carry arbitrary data. Practical uses include:

- **Store-and-forward messaging** — send text messages to operators who collect them later, across time zones or operating schedules
- **Image and file relay** — transmit JPEG images, PDFs, or data files to remote stations. A 100 KB image at even modest data rates transfers in seconds via the wideband transponder
- **Telemetry aggregation** — ground stations monitoring amateur satellites can bundle telemetry data and forward it to central servers via QO-100, providing an alternative delivery path when internet connectivity is unreliable
- **Emergency/resilient communications** — DTN's store-and-forward model means data survives infrastructure outages. If the internet is down, the space link still works
- **Network experiments** — test delay-tolerant protocols, routing algorithms, and convergence layer performance over a real space link

The protocol does not care what is inside the bundle. If it fits in the bandwidth, it can travel.

---

## How DTN Differs from DATV

DATV is real-time: you transmit video, someone watches it live. If nobody is watching, the transmission is lost.

DTN is asynchronous: you transmit data, it waits until someone is ready to receive it. Nothing is lost to timing. The network has memory.

Both can coexist on the wideband transponder. A DATV operator streaming live and a DTN operator sending stored bundles use the same bandwidth allocation approach — they just occupy different carrier slots. There is no interference or conflict, just shared transponder capacity managed by normal QO-100 band planning etiquette.

---

## Getting Started

The RADIANT software is freely available:

1. Install the RADIANT client on a Raspberry Pi or Linux machine
2. Choose a DTN engine (ION-DTN, Hardy, or µD3TN) — installation scripts provided
3. Configure your callsign as your DTN Endpoint Identifier (`dtn://yourcall/service`)
4. Point your dish at QO-100 and configure your RF chain as you would for DATV
5. Connect to the RADIANT network and begin exchanging bundles

The project provides documentation, configuration examples, and a growing community of operators across Europe testing the system. Phase 1 terrestrial testing is underway using 9600-baud packet radio; Phase 1.5 via QO-100 is the next milestone.

No encryption is used — all data is in the clear, fully compliant with amateur radio regulations. Every transmission carries your callsign in its addressing, satisfying station identification requirements automatically.

---

## The Bigger Picture

QO-100 is a stepping stone. RADIANT's roadmap progresses from terrestrial links through QO-100, to a LEO CubeSat, and ultimately to a cislunar node near the Moon. The protocols are the same at every phase — only the delay and link margin change.

For DATV operators, the wideband transponder provides the bandwidth to test these protocols at meaningful data rates over a real space link. Your ground station hardware is already capable. The missing piece is software — and that is free, open-source, and waiting to be downloaded.

If you fancy sending data via space that does not require anyone to be listening at the exact moment you transmit, DTN via QO-100 is worth a look.

**[FIGURE 3: Side-by-side comparison — left: traditional DATV (live stream, both stations active simultaneously); right: DTN (store-and-forward, asynchronous delivery). Caption: "DATV requires real-time viewers. DTN stores data until the recipient is ready — both via the same transponder."]**

---

**Contact:**
David Johnson, G4DPZ
Email: dave@g4dpz.me.uk
Website: https://radiant.amsat-uk.org
Source code: https://github.com/g4dpz/cislunar_proposal

---

*David Johnson, G4DPZ, is the project lead for RADIANT and Honorary Secretary of AMSAT-UK. The project welcomes DATV and data mode operators interested in space networking experiments via QO-100.*
