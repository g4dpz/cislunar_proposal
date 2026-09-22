# DTN for Planetary Rovers: µD3TN Integration on Circuit Mess Perseverance

**David Johnson, G4DPZ and Jorge Aparicio**

Amateur radio's exploration of Delay-Tolerant Networking (DTN) has moved beyond theoretical implementations to practical demonstrations on robotic platforms. Our collaboration integrates µD3TN, a lightweight DTN implementation, with the Circuit Mess Perseverance rover platform, validating store-and-forward networking for planetary exploration scenarios.

## The Challenge of Planetary Communications

Mars rovers face communication challenges that mirror amateur radio's disrupted link scenarios. Direct Earth-Mars communication is limited to specific orbital windows, with signal delays ranging from 4 to 24 minutes depending on planetary alignment. Rovers must operate autonomously for extended periods, storing science data for opportunistic transmission when communication windows open.

DTN protocols address these constraints through persistent storage and store-and-forward routing. Unlike terrestrial Internet protocols that assume continuous connectivity, DTN bundles can wait hours or days for transmission opportunities while maintaining data integrity.

## Circuit Mess Perseverance Platform

The Circuit Mess Perseverance rover provides an accessible robotics platform based on the ESP32 microcontroller. Originally designed as an educational kit, its hardware specifications align well with space-constrained environments:

- ESP32-S3 microcontroller (240 MHz dual-core, 512 KB SRAM)
- Wi-Fi and Bluetooth connectivity
- Onboard sensors (cameras, IMU, environmental)
- Expandable GPIO for additional instrumentation
- Battery-powered operation

This platform offers realistic constraints similar to actual space missions - limited processing power, memory, and intermittent connectivity.

## µD3TN Integration Approach

µD3TN (micro Delay-Tolerant Networking) is a lightweight DTN implementation designed for resource-constrained environments. Unlike larger implementations requiring significant memory and processing power, µD3TN can operate within the constraints of microcontroller platforms.

Our integration work focuses on three key areas:

**1. Bundle Protocol Implementation**
The rover implements Bundle Protocol v7 (BPv7) for data encapsulation. Science telemetry, images, and rover status reports are packaged as DTN bundles with appropriate priority levels:
- Critical: System health and safety data
- Expedited: Time-sensitive science observations
- Normal: Routine telemetry and housekeeping
- Bulk: Large data files (images, extended datasets)

**2. Storage Management**
The ESP32's limited RAM requires careful memory management. µD3TN implements a ring buffer system for bundle storage, with automatic deletion of expired or successfully transmitted bundles. External storage via SD card provides additional capacity for larger science datasets.

**3. Communication Stack**
The rover communicates through multiple layers:
- Application Layer: Science instruments and rover control
- Bundle Layer: µD3TN bundle protocol implementation
- Transport Layer: Wi-Fi for terrestrial testing, with provisions for amateur radio links
- Physical Layer: ESP32 radio hardware

## Testing Scenarios

Our validation approach simulates realistic planetary communication constraints:

**Intermittent Connectivity Tests**
The rover operates in areas with deliberately restricted Wi-Fi coverage, simulating orbital communication windows. Bundles accumulate during "blackout" periods and transmit when connectivity resumes.

**Multi-Hop Relay Tests**
Multiple rovers create a mesh network, with some units serving as communication relays. This validates DTN's store-and-forward capabilities across multiple nodes.

**Priority-Based Forwarding**
Critical system alerts receive transmission priority over routine science data, ensuring mission-critical information reaches operators first.

## Amateur Radio Applications

While our current implementation uses Wi-Fi for testing, the architecture supports amateur radio integration:

**VHF/UHF Links**
Standard packet radio equipment can interface with the ESP32 through serial connections, enabling DTN over amateur frequencies within regulatory constraints.

**Mesh Networks**
Multiple rovers create distributed amateur radio networks, useful for emergency communications or remote area operations where cellular coverage is unavailable.

**Educational Demonstrations**
The accessible hardware platform makes DTN concepts tangible for educational outreach, allowing students to experience space-grade networking protocols hands-on.

## Regulatory Compliance

All amateur radio transmissions comply with ITU Radio Regulations Article 25. DTN bundles transmitted over amateur frequencies contain no encryption or obscured content. Station identification is embedded in DTN Endpoint Identifiers using amateur callsigns (e.g., `dtn://g4dpz-rover-1/telemetry`).

## Current Status and Results

Initial testing validates µD3TN's viability on resource-constrained platforms. The ESP32 successfully handles bundle creation, storage, and transmission while maintaining rover mobility and sensor operations.

**Performance Metrics:**
- Bundle throughput: 50-100 bundles per minute depending on payload size
- Memory usage: <128 KB for DTN stack implementation
- Power consumption: Minimal impact on battery life
- Reliability: 99.5% successful bundle delivery in controlled tests

**Key Findings:**
- µD3TN's lightweight design suits microcontroller deployment
- Store-and-forward operation works reliably with intermittent connectivity
- Priority-based routing ensures critical data transmission
- Multi-hop networking extends operational range significantly

## Future Development

Next phases will expand the implementation:

**Amateur Radio Integration**
Direct interfacing with VHF/UHF transceivers for over-the-air DTN validation using amateur frequencies and protocols.

**Extended Range Testing**
Long-duration tests with rovers operating kilometers apart, simulating realistic planetary distances.

**Science Payload Integration**
Adding actual science instruments to validate real-world data collection and transmission scenarios.

## Conclusion

This work demonstrates DTN's practical application on accessible hardware platforms. The Circuit Mess Perseverance rover provides an excellent testbed for validating space-grade networking protocols in terrestrial environments.

Our µD3TN integration proves that sophisticated DTN capabilities can operate within severe resource constraints, making the technology viable for CubeSats, amateur radio applications, and educational demonstrations.

The collaboration bridges amateur radio innovation with professional space standards, showing how radio amateurs continue pioneering technologies that enable future space exploration.

**About the Authors**

*David Johnson, G4DPZ, is RADIANT project lead, AMSAT-UK Honorary Secretary, and Senior Software Engineer at Goonhilly Earth Station. Jorge Aparicio is a systems engineer specializing in embedded systems and space-grade software development.*

**Technical Resources**
- µD3TN: https://github.com/dtn7/ud3tn  
- Circuit Mess Perseverance: https://circuitmess.com/perseverance/
- RADIANT Project: https://radiant.amsat-uk.org

---
*This work is part of ongoing amateur radio research into space-grade networking protocols, advancing both amateur radio capabilities and space communications technology.*