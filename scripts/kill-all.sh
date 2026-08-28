#!/usr/bin/env bash
# Kill all local Optima stack processes.
#
# Targets:
#   - `go run ./cmd/<svc>` / `go run ./cmd/<svc>/main.go` (incl. ctrl -up)
#   - compiled children under /tmp/go-build* and go-build cache (cwd = repo)
#   - ad-hoc binaries /tmp/optima-*
#   - (optional) listeners on gateway/NATS ports
#
# Does NOT kill shells, editors, or node (web :5173) — only stack processes.
#
# Preferred entry (Task):  task down
#                          task down -- --dry-run
#                          task down:dry
#
# Direct:
#   bash scripts/kill-all.sh
#   bash scripts/kill-all.sh --dry-run
#   KILL_PORTS=0 bash scripts/kill-all.sh
#   WAIT_SECS=3 bash scripts/kill-all.sh
#
# Env:
#   DRY_RUN      1 = list only
#   KILL_PORTS   1 = free 8080/4222/8222 (default 1)
#   WAIT_SECS    seconds between SIGTERM and SIGKILL (default 2)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DRY_RUN="${DRY_RUN:-0}"
KILL_PORTS="${KILL_PORTS:-1}"
WAIT_SECS="${WAIT_SECS:-2}"

for arg in "$@"; do
  case "$arg" in
    --dry-run|-n) DRY_RUN=1 ;;
    --no-ports)   KILL_PORTS=0 ;;
    --force|-9)   WAIT_SECS=0 ;;
    -h|--help)
      sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "unknown arg: $arg (try --help)" >&2
      exit 2
      ;;
  esac
done

SERVICES=()
if [[ -d "$ROOT/services" ]]; then
  for d in "$ROOT/services"/*/; do
    [[ -d "$d" ]] || continue
    SERVICES+=("$(basename "$d")")
  done
fi
if ((${#SERVICES[@]} == 0)); then
  SERVICES=(nats gate logg ctrl)
fi

# Default ports (gateway + NATS client + NATS HTTP monitor).
PORTS=(8080 4222 8222)

SELF_PID=$$
SELF_PGID=$(ps -o pgid= -p "$SELF_PID" 2>/dev/null | tr -d ' ' || echo "")

declare -A PIDS=()

add_pid() {
  local pid="$1"
  [[ -n "$pid" ]] || return 0
  [[ "$pid" =~ ^[0-9]+$ ]] || return 0
  # never kill ourselves or our process group leader incorrectly
  if [[ "$pid" == "$SELF_PID" ]]; then
    return 0
  fi
  if [[ -n "$SELF_PGID" && "$pid" == "$SELF_PGID" ]]; then
    return 0
  fi
  # skip if already dead
  if ! kill -0 "$pid" 2>/dev/null; then
    return 0
  fi
  PIDS["$pid"]=1
}

proc_cmdline() {
  local pid="$1"
  if [[ -r "/proc/$pid/cmdline" ]]; then
    tr '\0' ' ' <"/proc/$pid/cmdline" 2>/dev/null || true
    return
  fi
  ps -o args= -p "$pid" 2>/dev/null || true
}

proc_cwd() {
  local pid="$1"
  readlink -f "/proc/$pid/cwd" 2>/dev/null || true
}

proc_exe() {
  local pid="$1"
  readlink -f "/proc/$pid/exe" 2>/dev/null || true
}

# 1) go run ./cmd/<svc> … (and main.go variants) for known services
for svc in "${SERVICES[@]}"; do
  while read -r pid; do
    add_pid "$pid"
  done < <(pgrep -f -- "go run (\./)?cmd/${svc}(/main\.go)?( |$)" 2>/dev/null || true)
  # service-local entrypoints: go run ./services/<svc>/cmd/...
  while read -r pid; do
    add_pid "$pid"
  done < <(pgrep -f -- "go run (\./)?services/${svc}/" 2>/dev/null || true)
done

# 2) /tmp/optima-* binaries (ad-hoc builds)
while read -r pid; do
  add_pid "$pid"
done < <(pgrep -f -- '/tmp/optima-' 2>/dev/null || true)

# 3) Compiled children: cwd is this repo, exe is go-build or /tmp/optima-*
#    Covers /tmp/go-build*/b001/exe/* and ~/.cache/go-build/*/* binaries.
for pid_dir in /proc/[0-9]*; do
  pid="${pid_dir#/proc/}"
  [[ "$pid" =~ ^[0-9]+$ ]] || continue
  cwd="$(proc_cwd "$pid")"
  [[ "$cwd" == "$ROOT" || "$cwd" == "$ROOT"/* ]] || continue
  exe="$(proc_exe "$pid")"
  cmd="$(proc_cmdline "$pid")"
  # go run wrappers already caught; also catch compiled exes + go tool children
  if [[ "$exe" == *"/go-build"* || "$exe" == /tmp/optima-* ]]; then
    add_pid "$pid"
    continue
  fi
  # cmdline still looks like stack start even if cwd match alone is weak
  if [[ "$cmd" == *"go run ./cmd/"* || "$cmd" == *"go run ./services/"* ]]; then
    add_pid "$pid"
  fi
done

# 4) Port holders (gateway / nats) — only if they look like stack procs
if [[ "$KILL_PORTS" == "1" || "$KILL_PORTS" == "true" || "$KILL_PORTS" == "yes" ]]; then
  for port in "${PORTS[@]}"; do
    # ss: users:(("name",pid=N,fd=F))
    if command -v ss >/dev/null 2>&1; then
      while read -r line; do
        if [[ "$line" =~ pid=([0-9]+) ]]; then
          pid="${BASH_REMATCH[1]}"
          exe="$(proc_exe "$pid")"
          cmd="$(proc_cmdline "$pid")"
          cwd="$(proc_cwd "$pid")"
          # only kill if related to optima stack
          if [[ "$cwd" == "$ROOT" || "$cwd" == "$ROOT"/* \
             || "$exe" == *"/go-build"* || "$exe" == /tmp/optima-* \
             || "$cmd" == *"go run ./cmd/"* \
             || "$cmd" == *"/tmp/optima-"* \
             || "$(basename "$exe" 2>/dev/null)" =~ ^(nats|gate|space|sheet|optima-.*)$ ]]; then
            add_pid "$pid"
          fi
        fi
      done < <(ss -tlnp 2>/dev/null | grep -E ":${port}\s" || true)
    elif command -v lsof >/dev/null 2>&1; then
      while read -r pid; do
        add_pid "$pid"
      done < <(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)
    elif command -v fuser >/dev/null 2>&1; then
      # fuser prints pids; capture carefully
      while read -r pid; do
        add_pid "$pid"
      done < <(fuser "${port}/tcp" 2>/dev/null | tr ' ' '\n' | grep -E '^[0-9]+$' || true)
    fi
  done
fi

count=${#PIDS[@]}
if (( count == 0 )); then
  echo "==> no Optima stack processes found"
  exit 0
fi

echo "==> found $count process(es) to stop (repo=$ROOT)"
sorted_pids=($(printf '%s\n' "${!PIDS[@]}" | sort -n))
for pid in "${sorted_pids[@]}"; do
  cmd="$(proc_cmdline "$pid")"
  cmd="${cmd//$'\n'/ }"
  # trim long lines
  if ((${#cmd} > 120)); then
    cmd="${cmd:0:117}..."
  fi
  printf '    pid=%-7s %s\n' "$pid" "${cmd:-?}"
done

if [[ "$DRY_RUN" == "1" || "$DRY_RUN" == "true" || "$DRY_RUN" == "yes" ]]; then
  echo "==> dry-run: not sending signals"
  exit 0
fi

echo "==> SIGTERM..."
for pid in "${sorted_pids[@]}"; do
  kill -TERM "$pid" 2>/dev/null || true
done

# also try process groups of go run parents (negative PGID) for leftover kids
for pid in "${sorted_pids[@]}"; do
  pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ' || true)
  if [[ -n "$pgid" && "$pgid" != "0" && "$pgid" != "$SELF_PGID" ]]; then
    kill -TERM -- "-$pgid" 2>/dev/null || true
  fi
done

deadline=$(( $(date +%s) + WAIT_SECS ))
while (( $(date +%s) < deadline )); do
  alive=0
  for pid in "${sorted_pids[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      alive=1
      break
    fi
  done
  (( alive == 0 )) && break
  sleep 0.2
done

still=()
for pid in "${sorted_pids[@]}"; do
  if kill -0 "$pid" 2>/dev/null; then
    still+=("$pid")
  fi
done

if ((${#still[@]} > 0)); then
  echo "==> SIGKILL ${#still[@]} stubborn process(es)..."
  for pid in "${still[@]}"; do
    kill -KILL "$pid" 2>/dev/null || true
    pgid=$(ps -o pgid= -p "$pid" 2>/dev/null | tr -d ' ' || true)
    if [[ -n "$pgid" && "$pgid" != "0" && "$pgid" != "$SELF_PGID" ]]; then
      kill -KILL -- "-$pgid" 2>/dev/null || true
    fi
  done
  sleep 0.2
fi

left=0
for pid in "${sorted_pids[@]}"; do
  if kill -0 "$pid" 2>/dev/null; then
    echo "    WARN: still alive pid=$pid $(proc_cmdline "$pid")"
    left=$((left + 1))
  fi
done

if (( left == 0 )); then
  echo "==> done — stack processes stopped"
else
  echo "==> done with $left still alive (may need sudo or different user)"
  exit 1
fi
