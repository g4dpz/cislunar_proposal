#!/bin/bash
# sync-specs-to-docs.sh — Copy spec files from .kiro/specs/ to docs/
#
# Run from the project root:
#   ./scripts/sync-specs-to-docs.sh
#
# Syncs requirements.md, design.md, and tasks.md for all mapped specs.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SPECS_DIR="$PROJECT_ROOT/.kiro/specs"
DOCS_DIR="$PROJECT_ROOT/docs"

# Map spec directories to docs directories
# Format: spec-name:docs-name (use same name if identical)
PHASE_MAP=(
  "terrestrial-dtn-phase1:terrestrial-dtn-phase1"
  "qo-100-geo-satellite-dtn:qo-100-geo-satellite-dtn"
  "cubesat-em-phase2:cubesat-em-phase2"
  "leo-cubesat-phase3:leo-cubesat-phase3"
  "cislunar-phase4:cislunar-phase4"
  "cislunar-amateur-dtn-payload:cislunar-amateur-dtn-payload"
  "multi-node-contact-graph:multi-node-contact-graph"
  "contact-log:contact-log"
  "test-framework-srs-sdd:test-framework-srs-sdd"
)

synced=0
skipped=0

for mapping in "${PHASE_MAP[@]}"; do
  spec_name="${mapping%%:*}"
  docs_name="${mapping##*:}"

  # Create target directory if it doesn't exist
  mkdir -p "$DOCS_DIR/$docs_name"

  # Sync requirements.md, design.md, and tasks.md
  for file in requirements.md design.md tasks.md; do
    src="$SPECS_DIR/$spec_name/$file"
    dst="$DOCS_DIR/$docs_name/$file"

    if [[ ! -f "$src" ]]; then
      continue
    fi

    # Only copy if source is newer or destination doesn't exist
    if [[ ! -f "$dst" ]] || [[ "$src" -nt "$dst" ]]; then
      cp "$src" "$dst"
      echo "  SYNC  $spec_name/$file → docs/$docs_name/$file"
      ((synced++)) || true
    else
      echo "  OK    $docs_name/$file (up to date)"
    fi
  done
done

echo ""
echo "Done: $synced synced, $skipped skipped."
