## Terrestrial Testing: µD3TN Rover Platform

Beyond the satellite missions, RADIANT's DTN protocols are being validated on terrestrial robotic platforms. Jorge Amodio and I are building a Circuit Mess Perseverance rover with integrated µD3TN and will be experimenting with a three-node DTN solution to communicate with the rover, demonstrating store-and-forward networking in mobile scenarios.

The ESP32-S3 based rover provides realistic space mission constraints - 240 MHz dual-core processing, 512 KB SRAM, and battery-powered operation. Originally an educational kit costing under £200, it offers amateur radio operators an accessible platform for DTN experimentation. The kit is available from Circuit Mess at https://circuitmess.com/gb-en/products/nasa-mars-perseverance-rover

Our µD3TN implementation packages telemetry and sensor data as DTN bundles with priority handling - critical system alerts transmit before routine housekeeping data. The ESP32's limited RAM requires careful bundle storage using ring buffers, with an SD card providing additional capacity for larger datasets.

We're testing intermittent connectivity scenarios where the rover operates in areas with restricted Wi-Fi coverage, simulating orbital communication windows. Bundles accumulate during blackout periods and transmit when connectivity returns. Our planned three-node DTN network will create mesh connectivity with store-and-forward capability, extending operational range through relay operations.

The architecture supports amateur radio integration via serial interfaces to VHF/UHF transceivers, enabling DTN over amateur frequencies. Station identification uses callsign-embedded DTN Endpoint Identifiers like `dtn://g4dpz-rover-1/telemetry`, ensuring ITU regulatory compliance.

This ground-based validation complements RADIANT's space missions, proving DTN protocols work reliably from terrestrial applications through cislunar operations.

The rover project demonstrates that space-grade DTN protocols can run on affordable hardware accessible to the amateur radio community. Amateur operators interested in DTN can start with similar ESP32 platforms, gradually adding radio interfaces for over-the-air testing using the open-source µD3TN software.