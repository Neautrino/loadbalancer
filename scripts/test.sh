#!/usr/bin/env bash
#
# End-to-end test: starts 3 backends + the LB, then exercises every balancing
# algorithm by patching config.yaml, restarting the LB, and reporting how many
# requests each backend served (with a bar chart + per-client mapping for the
# hash strategies). Cleans up (kills processes, restores config) on exit.
#
# Usage:  ./scripts/test.sh            # test all algorithms
#         ./scripts/test.sh weighted   # test just one
#
set -eu

LB="localhost:8080"
PORTS=(9001 9002 9003)
IPS=(1.1.1.1 2.2.2.2 3.3.3.3 4.4.4.4 5.5.5.5 6.6.6.6 7.7.7.7 8.8.8.8)
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ALGOS=(round_robin weighted least_conn ip_hash consistent_hash)
[ $# -ge 1 ] && ALGOS=("$1")

# ---- colours (fall back to plain if not a tty) ----
if [ -t 1 ]; then B="\033[1m"; DIM="\033[2m"; GRN="\033[32m"; CYN="\033[36m"; RST="\033[0m"
else B=""; DIM=""; GRN=""; CYN=""; RST=""; fi

echo "building binaries..."
go build -o ./bin/backend ./cmd/backend
go build -o ./bin/lb ./cmd/lb

CFG_BACKUP="$(mktemp)"
cp config.yaml "$CFG_BACKUP"

BACKEND_PIDS=()
LB_PID=""

cleanup() {
  echo
  echo "cleaning up..."
  [ -n "$LB_PID" ] && kill "$LB_PID" 2>/dev/null || true
  for pid in "${BACKEND_PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
  cp "$CFG_BACKUP" config.yaml && rm -f "$CFG_BACKUP"
}
trap cleanup EXIT

wait_lb() {
  for _ in $(seq 1 30); do
    curl -s -o /dev/null "$LB/" 2>/dev/null && return 0
    sleep 0.2
  done
  return 1
}
start_lb() { ./bin/lb >/dev/null 2>&1 & LB_PID=$!; wait_lb || { echo "!! LB failed to start"; exit 1; }; }
stop_lb()  { [ -n "$LB_PID" ] && kill "$LB_PID" 2>/dev/null || true; wait "$LB_PID" 2>/dev/null || true; LB_PID=""; }

describe() {
  case "$1" in
    round_robin)     echo "cycles through backends in order — expect an even split" ;;
    weighted)        echo "traffic proportional to weight (9001=3, 9002=1, 9003=1) — expect ~3:1:1" ;;
    least_conn)      echo "fewest in-flight requests (concurrent /slow) — expect balanced" ;;
    ip_hash)         echo "hash(client IP) % N — each IP always maps to the same backend" ;;
    consistent_hash) echo "hash ring + virtual nodes — sticky, minimal remap when membership changes" ;;
  esac
}

# reads "port NNNN" lines on stdin, prints a bar chart with counts + percentages
bars() {
  local input total c pct barlen bar
  input="$(cat)"
  total="$(printf '%s\n' "$input" | grep -c 'port' || true)"
  echo -e "  ${DIM}distribution over ${total} requests:${RST}"
  for p in "${PORTS[@]}"; do
    c="$(printf '%s\n' "$input" | grep -c "port ${p}" || true)"
    if [ "$total" -gt 0 ]; then pct=$(( c * 100 / total )); else pct=0; fi
    barlen=$(( pct / 4 ))
    bar="$(printf '%*s' "$barlen" '' | tr ' ' '#')"
    printf "    ${CYN}%s${RST} | ${GRN}%-25s${RST} %3d  (%d%%)\n" "$p" "$bar" "$c" "$pct"
  done
}

# start backends once; they stay up across all algorithm runs
for p in "${PORTS[@]}"; do
  ./bin/backend -port="$p" >/dev/null 2>&1 &
  BACKEND_PIDS+=($!)
done
sleep 0.5

for algo in "${ALGOS[@]}"; do
  sed -i -E "s/^algorithm:.*/algorithm: ${algo}/" config.yaml
  start_lb
  echo
  echo -e "${B}══════ algorithm: ${algo} ══════${RST}"
  echo -e "  ${DIM}$(describe "$algo")${RST}"

  start_ns=$(date +%s%N)
  case "$algo" in
    round_robin|weighted)
      for _ in $(seq 1 30); do curl -s "$LB/"; done | bars
      ;;
    least_conn)
      tmp="$(mktemp)"
      ( for _ in $(seq 1 12); do curl -s "$LB/slow" >>"$tmp" & done; wait )
      bars < "$tmp"; rm -f "$tmp"
      ;;
    ip_hash|consistent_hash)
      echo -e "  ${DIM}client IP -> backend (should be stable per IP):${RST}"
      tmp="$(mktemp)"
      for ip in "${IPS[@]}"; do
        port="$(curl -s -H "X-Forwarded-For: ${ip}" "$LB/" | grep -oE 'port [0-9]+')"
        printf "    %-9s ${CYN}->${RST} %s\n" "$ip" "$port"
        echo "$port" >> "$tmp"
      done
      bars < "$tmp"; rm -f "$tmp"
      # re-check stickiness: same IP twice must hit the same backend
      a="$(curl -s -H "X-Forwarded-For: 7.7.7.7" "$LB/" | grep -oE 'port [0-9]+')"
      b="$(curl -s -H "X-Forwarded-For: 7.7.7.7" "$LB/" | grep -oE 'port [0-9]+')"
      [ "$a" = "$b" ] && echo -e "  ${GRN}✓ sticky${RST}: 7.7.7.7 hit ${a} both times" \
                      || echo -e "  ✗ NOT sticky: ${a} then ${b}"
      ;;
  esac
  dur_ms=$(( ($(date +%s%N) - start_ns) / 1000000 ))
  echo -e "  ${DIM}took ${dur_ms}ms${RST}"

  stop_lb
done

echo
echo -e "${GRN}done — all requested algorithms tested.${RST}"
