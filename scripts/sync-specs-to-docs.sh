#!/bin/bash
#
# Sync spec files from .kiro/specs/ to docs/ directory
# Excludes .config.kiro files (Kiro internal configuration)
#

set -e

echo "Syncing spec files from .kiro/specs/ to docs/..."
echo ""

# Array of spec directories to sync
SPECS=(
    "terrestrial-dtn-phase1"
    "qo-100-geo-satellite-dtn"
    "cubesat-em-phase2"
    "leo-cubesat-phase3"
    "cislunar-phase4"
    "cislunar-amateur-dtn-payload"
)

# Sync each spec directory
for spec in "${SPECS[@]}"; do
    if [ -d ".kiro/specs/$spec" ]; then
        echo "Syncing $spec..."
        rsync -av --delete --exclude='.config.kiro' \
            ".kiro/specs/$spec/" "docs/$spec/"
    else
        echo "Warning: .kiro/specs/$spec not found, skipping"
    fi
done

echo ""
echo "Sync complete!"
echo ""
echo "Summary:"
for spec in "${SPECS[@]}"; do
    if [ -d "docs/$spec" ]; then
        file_count=$(find "docs/$spec" -type f | wc -l | tr -d ' ')
        echo "  - $spec: $file_count files"
    fi
done
