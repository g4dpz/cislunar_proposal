#!/bin/bash
# Sync spec documents from .kiro/specs/ to docs/ for the public repository.
# Run this before committing to keep docs/ up to date.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

KIRO_SPECS="$PROJECT_DIR/.kiro/specs"
DOCS_DIR="$PROJECT_DIR/docs"

if [ ! -d "$KIRO_SPECS" ]; then
    echo "No .kiro/specs/ directory found."
    exit 1
fi

echo "Syncing specs to docs/..."

# Sync each spec directory, copying only .md files
for spec_dir in "$KIRO_SPECS"/*/; do
    spec_name=$(basename "$spec_dir")
    target="$DOCS_DIR/$spec_name"
    mkdir -p "$target"
    rsync -av --include='*.md' --exclude='*' --exclude='.config.kiro' "$spec_dir" "$target/"
done

echo "Done. docs/ is up to date."
echo ""
echo "Synced specs:"
find "$DOCS_DIR" -name "*.md" | sort | sed "s|$PROJECT_DIR/||"
