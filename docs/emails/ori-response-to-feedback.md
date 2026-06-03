# Draft Response to Michelle's Feedback

**To:** Michelle Thompson (W5NYV)  
**From:** Dave Johnson (G4DPZ)  
**Subject:** Re: RADIANT feedback — fair points

---

Hi Michelle,

Thanks for taking the time to look at this and for the honest assessment. You've raised several points that are fair and I want to address them directly.

**On the "Earth-Moon / Earth-Mars demonstrations" language** — you're right, that's overclaiming. It's a lab simulation with injected propagation delay, not an RF path to the Moon. The DTN behaviour is real (store-and-forward across disconnected hops with the correct timing), but calling it a "demonstration" without qualifying that it's simulated delay over a terrestrial link is misleading. I'll fix the wording on the site.

**On the marketing tone** — guilty as charged. The site was written partly to attract collaborators and partly for an ESA ARTES expression of interest, and it ended up reading like the project is further along than it actually is. Phase 1 is two Raspberry Pis, two TNC4s, and two FT-817s doing ping and file transfer over VHF. That's it. I'll make the current status vs. roadmap distinction much clearer.

**On the repo** — it's currently private while we sort out some contribution agreements with AMSAT-UK. That's no excuse for implying it's available when it isn't. I'll either make it public or clearly state it's in development.

**On the PHY layer** — completely agree there's no modulation or coding innovation in Phase 1. That was a deliberate choice: get the DTN stack working end-to-end over the simplest possible link, then improve the PHY. The Phase 2 engineering model moves to software-defined IQ baseband on an STM32U585 with the intention of adding FEC (convolutional initially, LDPC for later phases) and adaptive modulation (4FSK, eventually QPSK). But none of that exists yet, and the site shouldn't imply otherwise.

The honest value proposition of RADIANT is: DTN integration work on amateur radio, done carefully with regulatory compliance baked in, building toward a flight opportunity. It's not a PHY innovation project — it's a networking-layer project that happens to need a competent PHY underneath it. Which is exactly where I think ORI's work could complement it.

I'd still be keen to have that conversation if you're open to it. Your DVB-S2 signal chain and LDPC work is solving the exact problem we'll face when we move beyond 9600 baud — and I'd rather adopt proven open-source work than build something inferior from scratch.

73,  
Dave, G4DPZ
