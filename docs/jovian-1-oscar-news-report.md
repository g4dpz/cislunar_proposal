# FUNcube Lite Payload Status for Jovian-1 Mission

**Project Background**

Jovian-1 is a 6U CubeSat being built by Space South Central, the UK's largest regional space cluster. The satellite is part of JUPITER (Joint Universities Programme for In-Orbit Training, Education and Research), involving the universities of Surrey, Portsmouth and Southampton.

AMSAT-UK is providing a FUNcube Lite payload with a mode U/V FM transponder for this mission. The payload will collect and transmit telemetry from Jovian-1 sub-systems for educational outreach, using the FUNcube data format.

**Payload Function**

The FUNcube Lite payload will gather telemetry from its radiation sensor and GPS information from the satellite's CAN bus. This data will map radiation throughout the orbit, identifying radiation concentration in polar regions and the South Atlantic Anomaly.

Jovian-1 uses commercial S and X band frequencies for primary communications through Surrey University's ground station. The FUNcube Lite payload operates on amateur UHF and VHF spectrum. When not transmitting telemetry, the payload operates as a mode U/V FM voice transponder.

**Current Status Report**

The FUNcube team has delivered the deployable antenna system and associated RF and data cables for their Jovian-1 payload in mid-2025. The hardware is currently either in the clean room or stored in its Pelican case awaiting integration. 

The remaining four flight boards have all been assembled and tested as far as possible. Of the four boards, two are dedicated to RF communication, while the other two boards — Command Control and Telemetry (CCT) — are microprocessor and CPLD based. The CCT boards require programming to communicate with the Jovian-1 payload processor via CAN bus.

## RF Communications Progress

The two RF communications boards completed their laboratory and acceptance testing in June 2025. In July 2025, the team conducted field testing at Dunstable Downs in Bedfordshire. They ran a series of tests with two ground stations located 25km and 60km away. 

The tests were successful, with data gathered to verify performance predictions and extrapolate to the >1,500km range required from orbit. Both RF boards have been cleaned and any component over 1 gram has been secured in preparation for integration and environmental testing.

## CCT Board Development Status

The CCT/CAN interface flight boards have been assembled and limited testing has been undertaken to simulate communication via CAN bus. Work on these boards was suspended at the end of November 2024 pending completion of satellite system software.

The CCT boards require programming to communicate with the Jovian-1 payload processor. The software for the real-time operating system on the satellite's main computer and the software for the payload processor on Jovian-1 remain under development. The CCT boards cannot be completed until these system interfaces are defined and operational.

## System Integration Dependencies

The FUNcube payload faces unique integration challenges due to its position within the satellite's data architecture. The payload operates at the end of the data chain, requiring input from multiple satellite subsystems before it can function.

This contrasts with instruments that generate data independently. The radiation payload led by Dr. Clewer, for example, operates as a standalone instrument. It gathers and stores measurement data in specific memory locations when powered on. Such instruments can be integrated without requiring detailed knowledge of other satellite subsystems.

The FUNcube payload must make requests to each satellite subsystem via the payload processor to collect data from sources such as:
- GPS systems
- Electrical power system telemetry  
- Solar panel data
- Reaction wheel status
- Other satellite subsystems

This data must flow from each subsystem to the Jovian-1 computer, then to the payload processor, before the FUNcube payload can access it via CAN bus at specific memory addresses. The payload's functionality depends on this complete data chain being operational.

## Completion Timeline

The CCT development team estimates that once the required system interface information becomes available, approximately three months will be needed to program and test the payload. If a simulator is not available, some of this work will need to take place at the University. The payload would need to be connected to Jovian-1 via CAN bus for final integration testing.

The FUNcube team continues to work closely with the Jovian-1 mission planners to resolve these integration challenges and complete their contribution to this important educational satellite mission.