# Reply to Leigh Torgerson

**To:** Leigh Torgerson (K6JLT)  
**From:** Dave Johnson (G4DPZ)  
**Subject:** Re: RADIANT — welcome aboard!

---

Hi Leigh,

Great to hear from you, and welcome aboard! Really pleased you're interested in RADIANT — your background with the ION testbed and DSN compatibility testing is exactly the kind of experience that would be invaluable to what we're building.

A bit about me — I'm Honorary Secretary of AMSAT-UK, a member of the ARISS-UK team, and am part of the FUNcube-1 mission team, so well connected on the amateur satellite side of things. Matt Cosby is actually my boss at Goonhilly! I'm based out of the Farnborough office, so if you're heading to the airshow on the 24th we're practically neighbours that week. I'd love to meet in person if your schedule allows — Farnborough or London work for me. And if you do make it down to Goonhilly I can help arrange that too.

Your QST article experience made me laugh — we've had similar timing frustrations. The pace of DTN development right now means anything written has a short shelf life. That said, I'm currently writing articles for AMSAT-UK OSCAR News and the RSGB Radio Communication magazine, and I'm just putting together a technical paper for the IPNSIG TD WG Zotero library covering the RADIANT architecture.

The Internet testbed Jorge is putting together sounds like a natural complement to what we're doing over RF. One of our key goals is DTN implementation interoperability — we want to validate that ION, µD3TN, ESA-DTN, and Hardy can all exchange bundles correctly over real links. Your experience running the JPL testbed and doing formal compatibility testing is directly relevant to that. We'd love your input on how to structure an interop test plan.

I'm particularly interested in Hardy — the Rust BPv7 implementation with `no_std` core libraries. For our flight hardware (STM32U585, 786 KB SRAM), Rust's memory safety guarantees and the ability to run without a full OS make it a compelling candidate. µD3TN is also high on the list for flight — it's already space-tested and designed for exactly the kind of constrained POSIX/bare-metal environment we're targeting. I'd be curious to hear your thoughts on how both compare to ION for constrained platforms, and whether you've seen either exercised against ION in interop testing.

Looking forward to working together. Let me know your UK dates and I'll sort out a meet-up.

73,
Dave, G4DPZ
