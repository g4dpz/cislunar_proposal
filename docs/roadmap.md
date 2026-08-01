# RADIANT Short-Term Roadmap

*Updated 27 July 2026*

## Done ✓

- EMF Camp (16-19 July) — village operational, lightning talk delivered
- radiant-ion deployed and running on Ubuntu VM
- ION-DTN lifecycle management working (ion-ctl.sh)
- LTP interop test passing (ION → Hardy, 1MB)
- TCPCLv3 inbound gateway configured and tested
- Three timing scenarios validated (delayed uplink, immediate, delayed downlink)
- Repo consolidated (code + website + docs)
- Website updated (architecture docs, SVG diagrams, conops, roadmap pages)
- EMF talk outline and slides prepared (40-min + 5-min ARISS lightning talk)
- AMSAT-BB post drafted
- Collaboration emails drafted (Lars, Loïc)

## This Week (27 July — 3 August)

- [ ] Send reply to Lars and Loïc
- [ ] Post-EMF write-up for AMSAT-UK website / social media
- [ ] Fix `radiant-dtn-abstraction` process spawning bug (ionadmin timeout)
- [ ] Integration call with Lars (UHF station as DTN node)

## August — Articles & Collaboration

### Magazine Articles (Tier 1 — submit August)

| Publication | Angle | Status |
|-------------|-------|--------|
| RadCom (RSGB) | Solar System network for general amateur readership | Full draft complete |
| OSCAR News (AMSAT-UK) | Cislunar DTN pathway, phased roadmap | Outline complete |
| AMSAT Journal (AMSAT-NA) | NASA protocol heritage, ION-DTN, interoperability | Outline complete |
| AMSAT-BB | Project announcement | Draft complete, ready to post |

### Collaboration

- [ ] Lars: integration call, define thesis topics, UHF station test plan
- [ ] Loïc: Hardy BPA on 44net, FOSM-1 as DTN endpoint, QO-100 setup progress
- [ ] Second ground station node online (Lars or Loïc)
- [ ] Matt Cosby / Goonhilly: project update, flat-sat concept

## September — October — Outreach & Talks

### Magazine Articles (Tier 2)

| Publication | Angle | Status |
|-------------|-------|--------|
| AMSAT-DL Journal | European ground segment, QO-100 | Outline complete |
| CQ-DATV | QO-100 wideband digital data | Not yet started |
| GRCon (GNU Radio Conference) | SDR modem / KISS link service for QO-100 DTN | Talk submission (check CFP) |
| IPNSIG | RADIANT as amateur implementation of SSI governance model | Presentation to Jorge / next meeting |

### Technical

- [ ] KISS link service adapter for QO-100 narrowband DTN demo
- [ ] Multi-station interop test (Lars UHF + Loïc Hardy + G4DPZ ION)
- [ ] Community dashboard for ground segment statistics
- [ ] AMSAT-UK Colloquium presentation (if scheduled)
- [ ] Flat-sat build (2x Pi + netem link simulator) for outreach demos
- [ ] Formal thesis topic proposals finalised with Lars

## November — December

### Magazine Articles (Tier 3 — when results available)

| Publication | Angle | Status |
|-------------|-------|--------|
| IEEE Aerospace Conference (short paper) | Backend-agnostic DTN architecture, LTP interop results | Needs QO-100 or multi-station results |
| SpaceOps / IAC Small Satellite | Ground segment automation, CGR validation | Needs Phase 1.5 results |
| LWN (Linux Weekly News) | Rust DTN tooling, open-source space networking | When tooling story more complete |

### Technical

- [ ] Phase 1.5 QO-100 DTN demonstration (if KISS adapter ready)
- [ ] CubeSat EM scoping (STM32U585 + IQ transceiver, Phase 2 prep)
- [ ] Year-end progress report for partners

## Publication Summary

| # | Publication | Audience | Length | Status |
|---|-------------|----------|--------|--------|
| 1 | RadCom (RSGB) | UK amateurs (general) | ~3,500 words | Draft complete |
| 2 | OSCAR News (AMSAT-UK) | Satellite operators | ~1,400 words | Outline complete |
| 3 | AMSAT Journal (AMSAT-NA) | US satellite community | ~2,000 words | Outline complete |
| 4 | AMSAT-BB | Global amateur email list | ~500 words | Draft complete |
| 5 | AMSAT-DL Journal | European satellite community | ~1,500 words | Outline complete |
| 6 | CQ-DATV | Digital ATV community | ~1,000 words | Not started |
| 7 | GRCon | SDR / GNU Radio community | Talk | Not started |
| 8 | IPNSIG | Solar System Internet community | Presentation | Contact Jorge |
| 9 | IEEE Aerospace | Academic / space engineering | Short paper | Needs results |
| 10 | SpaceOps / IAC | Mission operations | Paper | Needs results |
| 11 | LWN | Open-source developers | Article | Needs tooling story |
