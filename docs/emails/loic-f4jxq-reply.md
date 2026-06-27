# Reply to Loïc F4JXQ

**To:** Loïc (F4JXQ)  
**From:** Dave Johnson (G4DPZ)  
**Subject:** Re: RADIANT project — welcome!

---

Hi Loïc,

Thanks for reaching out — no need to apologise for a long email, this is exactly the kind of introduction I love to receive. Your background and current work align very well with what we're building.

A bit about me — I'm Honorary Secretary of AMSAT-UK and part of the FUNcube team, so I'm well connected on the amateur satellite side. RADIANT is supported by AMSAT-UK, AMSAT-DL, and Goonhilly Earth Station.

FOSM-1 is impressive — having a protocol experimentation payload already in orbit is a significant achievement. The Spino radio board being flight-proven and open-source is interesting too. We're still in the process of selecting flight hardware for Phase 3 (LEO CubeSat), so understanding what's already been demonstrated in orbit is valuable.

A few things that stand out from your email:

**Hardy testing** — you're one of very few people I know who are actively running Hardy. We've developed an LTP convergence layer adapter for Hardy (`hardy-ltp-cla`) and are building a DTN abstraction layer in Rust that supports both ION-DTN and Hardy as backend engines. One of our key goals is proving interoperability between the two. Having someone who knows Hardy's behaviour from direct experience would be extremely useful for that work. If you've found quirks, bugs, or undocumented behaviour, we'd want to know about it.

**44net Hardy BPA node** — that's great. Once you have that running, we could potentially do DTN bundle exchange testing between your node and ours over IP as a precursor to doing it over RF. That directly validates the abstraction layer's Hardy adapter.

**QO-100** — we're planning Phase 1.5 (DTN over the Es'hail-2 narrowband transponder) and it would be good to have another station working toward the same goal. When you get your setup assembled, let me know.

**Protocol verification background** — your PhD and startup experience in network protocol verification is directly relevant. We're using property-based testing (proptest in Rust) to formally verify correctness properties of our DTN stack. If formal verification of DTN protocol behaviour is something you'd enjoy contributing to, there's plenty of scope.

**FOSM-1 as a DTN platform** — longer term, if FOSM-1's protocol experimentation capability could support a BPv7/LTP experiment, that would be an extraordinary opportunity. A DTN bundle delivered via an amateur satellite that's already in orbit — even as a technology demonstration — would be a significant milestone for the project. Obviously that depends on FOSM-1's commissioning progress and available capacity, but it's worth keeping in mind.

Some concrete ways you could get involved right now:

1. **Hardy interop testing** — once your 44net node is up, we can attempt bundle exchange between your Hardy instance and our ION-DTN instance. That's exactly the interoperability data we need.
2. **Abstraction layer feedback** — we're building the DTN abstraction layer in Rust and the Hardy adapter is actively being developed. If you'd like to review or contribute to the Hardy adapter code, you'd be very welcome.
3. **QO-100 coordination** — when you're QRV, let's plan a DTN test over the transponder.
4. **Property-based testing** — if you're interested in applying your protocol verification experience to DTN correctness properties, we have a framework ready for contributions.

I'll add you to our contributor list. Would you like to join our next informal call with the other collaborators? We have Leigh Torgerson (ex-JPL, ION-DTN testbed lead), Jorge Amodio (IPNSIG), and connections into AMSAT-UK.

Looking forward to working together.

73,
Dave, G4DPZ
