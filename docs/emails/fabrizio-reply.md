# Reply to Fabrizio

**To:** Fabrizio  
**From:** Dave Johnson (G4DPZ)  
**Subject:** Re: RADIANT — contact plan and run evidence

---

Hi Fabrizio,

Thanks for the thoughtful email — you've put your finger on exactly the right question. And the good news is: we've already designed a system for precisely what you're describing.

We have a spec called the **Contact Log** — a versioned contact-plan and run-evidence logging system designed to enable cross-phase comparison of DTN performance. It's specced but not yet implemented, so your timing is excellent.

Here's the short version of what it does:

**Each DTN session produces an immutable, schema-versioned JSON record** containing:

1. **Contact Plan Snapshot** (what was planned):
   - Expected contact window start/end times
   - Link type and frequency band
   - Bitrate, modem type, modulation, framing assumptions
   - Node roles and DTN EIDs
   - Storage limits
   - Priority class distribution of queued bundles
   - Retransmission/custody assumptions

2. **Run Evidence** (what actually happened):
   - Actual link establishment and teardown times
   - Bytes transferred, bundles sent/received/delivered
   - Delivery latency, goodput, plan adherence ratio
   - Bundle IDs and routes selected
   - Delivery status and timing evidence per bundle
   - Interruption reasons if the session was cut short

3. **Phase Metadata** (context):
   - Which phase (terrestrial, QO-100, LEO, cislunar)
   - OWLT, orbital parameters if applicable
   - Modulation and coding scheme

The key design principle is that **every entry is self-describing** — you can take a log entry from Phase 1 terrestrial testing and directly compare normalised metrics (goodput, plan adherence, delivery success ratio) against a Phase 4 cislunar entry. The schema is versioned so entries remain interpretable as the system evolves.

To answer your direct question: we're treating contact plans and run evidence as **first-class versioned artifacts**, not just runtime config/logs. The plan is that each session produces a structured record that lives in a queryable store — queryable by phase, time range, node pair, and outcome. Machine-readable JSON with consistent field ordering.

Your SatNOGS/SSDV comparison is apt — the operational pattern is indeed similar. In our case, the Contact Log preserves exactly the context, ordering, timing, and evidence you need to make delayed/disrupted data useful after the fact.

If this is an area you'd like to contribute to, the spec is ready for implementation. We're building in Rust and the Contact Log integrates with the DTN abstraction layer we've just completed (which manages both ION-DTN and Hardy as backend engines). Your background in structured logging and operational evidence seems well-matched.

73,
Dave, G4DPZ
