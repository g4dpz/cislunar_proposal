# Draft Email: ORI Collaboration

**To:** Michelle Thompson (W5NYV)  
**From:** Dave Johnson (G4DPZ)  
**Subject:** RADIANT + ORI — possible overlap on open-source cislunar comms?

---

Hi Michelle,

Hope you're well. I've been following ORI's work on the DVB-S2 signal chain and the Opulent Voice modem — the Friedrichshafen demo looks impressive.

I'm writing because there's quite a bit of overlap between what you're building and a project I'm leading called RADIANT (Radio Amateur Delay-tolerant Interplanetary Networking Testbed). We're working with AMSAT-UK and AMSAT-DL on bringing DTN (Bundle Protocol v7 / LTP) to amateur radio, phased from terrestrial validation through LEO CubeSat to eventually cislunar links.

Our Phase 1 is operational — two-node terrestrial DTN over UHF using LTP-over-KISS at 9600 baud. We're now planning Phase 1.5 (QO-100 narrowband), Phase 2 (CubeSat engineering model on STM32U585), and eventually Phase 4 (cislunar S-band/X-band). It's the later phases where I think there could be interesting synergy with ORI's work:

- Your LDPC/BCH encoder and dvb_fpga RTL could be exactly what we need for higher-rate cislunar links where our current GFSK waveform won't cut it
- The pluto_msk uplink modem and opv-cxx-demod are solving problems we'll need to solve for Phase 3/4 RF front-ends
- We're both working toward open-source, regulatory-compliant amateur payloads for beyond-GEO

I'd be interested in having a chat to see if there's useful overlap — whether that's shared waveform development, test infrastructure, or just keeping each other informed so we're not duplicating effort. No specific proposal at this stage, just exploring where the Venn diagram sits.

Happy to jump on a call or meet at a conference if timing works. Our project site is https://radiant.amsat-uk.org if you want a look at what we're up to.

73,  
Dave, G4DPZ
