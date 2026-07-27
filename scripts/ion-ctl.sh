#!/bin/bash
# ion-ctl — Concrete ION-DTN controller
# Drives ION admin tools directly using pre-prepared config scripts.
#
# Usage:
#   ion-ctl.sh start <config-dir>
#   ion-ctl.sh stop
#   ion-ctl.sh status
#   ion-ctl.sh restart <config-dir>
#
# The config directory must contain: *.ionrc, *.ltprc, *.bprc, *.ipnrc

set -e

# ─── Helpers ──────────────────────────────────────────────────────────────────

die() { echo "ERROR: $*" >&2; exit 1; }

find_file() {
    local dir="$1" ext="$2"
    local f
    f=$(find "$dir" -maxdepth 1 -name "*.$ext" | head -1)
    [ -n "$f" ] || die "No .$ext file found in $dir"
    echo "$f"
}

check_bin() {
    command -v "$1" &>/dev/null || die "$1 not found on PATH"
}

# ─── Commands ─────────────────────────────────────────────────────────────────

cmd_start() {
    local config_dir="$1"
    [ -d "$config_dir" ] || die "Config directory not found: $config_dir"

    local ionrc ltprc bprc ipnrc
    ionrc=$(find_file "$config_dir" "ionrc")
    ltprc=$(find_file "$config_dir" "ltprc")
    bprc=$(find_file "$config_dir" "bprc")
    ipnrc=$(find_file "$config_dir" "ipnrc")

    check_bin ionadmin
    check_bin ltpadmin
    check_bin bpadmin
    check_bin ipnadmin

    echo "Starting ION from: $config_dir"

    # Run from the config directory — ION uses cwd for SDR/semaphore paths
    cd "$config_dir"

    echo "  ionadmin $(basename "$ionrc")..."
    ionadmin "$(basename "$ionrc")" >/dev/null 2>&1
    sleep 1

    echo "  ltpadmin $(basename "$ltprc")..."
    ltpadmin "$(basename "$ltprc")" >/dev/null 2>&1
    sleep 1

    echo "  bpadmin $(basename "$bprc")..."
    bpadmin "$(basename "$bprc")" >/dev/null 2>&1
    sleep 1

    echo "  ipnadmin $(basename "$ipnrc")..."
    ipnadmin "$(basename "$ipnrc")" >/dev/null 2>&1

    # Verify
    sleep 1
    if pgrep -x rfxclock >/dev/null 2>&1; then
        echo "ION started successfully."
        exit 0
    else
        die "ION failed to start (rfxclock not running)"
    fi
}

cmd_stop() {
    echo "Stopping ION..."
    check_bin ionstop

    ionstop 2>/dev/null || true
    sleep 2

    if command -v killm &>/dev/null; then
        killm 2>/dev/null || true
    fi

    echo "ION stopped."
}

cmd_status() {
    echo "=== ION Status ==="

    if pgrep -x rfxclock >/dev/null 2>&1; then
        echo "  Running: yes"
        echo ""
        echo "  Processes:"
        ps -eo pid,comm | grep -E "rfxclock|ltpclock|bpclock|udplso|udplsi|ltpdeliv|ltpmeter|bptransit|ipnadminep|ipnfw" | sed 's/^/    /'
    else
        echo "  Running: no"
    fi

    echo ""

    # Bundle stats if available
    if command -v bpstats &>/dev/null && pgrep -x rfxclock >/dev/null 2>&1; then
        echo "=== Bundle Statistics ==="
        bpstats 2>/dev/null | grep -v "^Stopping" | sed 's/^/  /' || echo "  (unavailable)"
        echo ""
    fi
}

cmd_restart() {
    local config_dir="$1"
    [ -d "$config_dir" ] || die "Config directory not found: $config_dir"

    echo "Restarting ION..."
    cmd_stop 2>/dev/null || true
    sleep 1
    cmd_start "$config_dir"
}

# ─── Main ─────────────────────────────────────────────────────────────────────

case "${1:-}" in
    start)
        [ -n "${2:-}" ] || die "Usage: $0 start <config-dir>"
        cmd_start "$(cd "$2" && pwd)"
        ;;
    stop)
        cmd_stop
        ;;
    status)
        cmd_status
        ;;
    restart)
        [ -n "${2:-}" ] || die "Usage: $0 restart <config-dir>"
        cmd_restart "$(cd "$2" && pwd)"
        ;;
    *)
        echo "ion-ctl — ION-DTN lifecycle controller"
        echo ""
        echo "Usage:"
        echo "  $0 start <config-dir>    Start ION with pre-prepared admin scripts"
        echo "  $0 stop                  Stop ION and clean up shared memory"
        echo "  $0 status                Show ION running state and processes"
        echo "  $0 restart <config-dir>  Stop then start ION"
        echo ""
        echo "Config directory must contain: *.ionrc, *.ltprc, *.bprc, *.ipnrc"
        exit 1
        ;;
esac
