#!/bin/bash
# sync-specs-to-docs.sh — Copy requirements.md from .kiro/specs/ to docs/
#
# Run from the project root:
#   ./scripts/sync-specs-to-docs.sh
#
# Only syncs requirements.md (not design.md or tasks.md which are internal).

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
)

synced=0
skipped=0

for mapping in "${PHASE_MAP[@]}"; do
  spec_name="${mapping%%:*}"
  docs_name="${mapping##*:}"

  src="$SPECS_DIR/$spec_name/requirements.md"
  dst="$DOCS_DIR/$docs_name/requirements.md"

  if [[ ! -f "$src" ]]; then
    echo "  SKIP  $spec_name (no requirements.md in specs)"
    ((skipped++))
    continue
  fi

  # Create target directory if it doesn't exist
  mkdir -p "$DOCS_DIR/$docs_name"

  # Only copy if source is newer or destination doesn't exist
  if [[ ! -f "$dst" ]] || [[ "$src" -nt "$dst" ]]; then
    cp "$src" "$dst"
    echo "  SYNC  $spec_name → docs/$docs_name/requirements.md"
    ((synced++))
  else
    echo "  OK    $docs_name/requirements.md (up to date)"
  fi
done

echo ""
echo "Done: $synced synced, $skipped skipped."
