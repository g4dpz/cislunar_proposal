# radiant-ion Deployment

## Prerequisites

- Ubuntu 22.04+ (or any Linux x86_64)
- ION-DTN 4.x on PATH
- Rust toolchain (`curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`)

## Package and Deploy

On your dev machine:

```bash
bash deploy/radiant-ion/package.sh
scp dist/radiant-ion-*.tar.gz user@server:/tmp/
```

On the VM:

```bash
mkdir -p ~/radiant-ion && cd ~/radiant-ion
tar xzf /tmp/radiant-ion-*.tar.gz
bash deploy/radiant-ion/install.sh
```

## Running

```bash
# Start ION + HTTP API
./bin/radiant-ion serve deploy/radiant-ion/config.yaml --port 3000

# Or just start ION (no API)
./bin/radiant-ion start deploy/radiant-ion/config.yaml

# Check status
./bin/radiant-ion status

# Stop
./bin/radiant-ion stop
```

## Rebuild After Source Changes

```bash
cd radiant-ion
cargo build --release
cp target/release/radiant-ion ../bin/
```

## Amateur Radio Compliance

No encryption. All payloads in the clear per ITU Article 25.
