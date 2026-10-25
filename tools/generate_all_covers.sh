#!/usr/bin/env bash
# Overnight run: every cover image the site uses, with FLUX.1-schnell.
#
#   tools/generate_all_covers.sh            # run everything (resumes where it left off)
#   tools/generate_all_covers.sh --dry-run  # just count what's left
#
# Steps, in order of how visible they are on the site:
#   1. replace the 100 report covers in place     (static/img/report-covers/<slug>.webp)
#   2. report cover pool, 12 categories × 118     (static/img/report-covers/pool/) — reports
#      generated beyond the first 100 had no cover at all before this
#   3. compliance course covers, 6 per topic       (static/img/compliance-covers/)
#
# Every step skips images that already exist (the report-cover step tracks
# replaced ones in its manifest), so it is safe to Ctrl-C, kill, or re-run.
# Each step is retried up to 3 times if it exits with an error. Ctrl-C stops
# the whole run (it used to end only the current step). The Mac is
# kept awake with caffeinate. Output goes to logs/covers-<timestamp>.log.
#
# Afterwards:
#   docker compose -f docker-compose-local.yml up -d --build acpwb_go   # embeds the compliance cover list
#   docker compose -f docker-compose-local.yml exec web python manage.py collectstatic --noinput
#   docker compose -f docker-compose-local.yml restart web              # reloads the report cover pool
# Originals of the replaced report covers are in git:
#   git checkout -- acpwb/static/img/report-covers
set -u
cd "$(dirname "$0")/.."

PY=.direnv/python-3.14/bin/python
[ -x "$PY" ] || PY=python3
SCRIPT=tools/generate_compliance_covers.py
mkdir -p logs
LOG="logs/covers-$(date +%Y%m%d-%H%M%S).log"

if [ "${1:-}" = "--dry-run" ]; then
  for mode in --report-covers --report-pool ""; do
    "$PY" "$SCRIPT" $mode --dry-run 2>/dev/null | head -1
  done
  exit 0
fi

# Ctrl-C (or kill) stops the whole run, not just the current step.
stop() {
  echo "=== $(date '+%F %T') stopped; re-run to resume" | tee -a "$LOG"
  exit 130
}
trap stop INT TERM

run_step() {
  local label=$1; shift
  local status
  for attempt in 1 2 3; do
    echo "=== $(date '+%F %T') $label (attempt $attempt)" | tee -a "$LOG"
    caffeinate -dims "$PY" "$SCRIPT" "$@" >>"$LOG" 2>&1
    status=$?
    if [ $status -eq 0 ]; then
      tail -1 "$LOG"
      return 0
    fi
    # 130/137/143: interrupted or killed, so stop rather than retry.
    if [ $status -eq 130 ] || [ $status -eq 137 ] || [ $status -eq 143 ]; then
      stop
    fi
    echo "!!! $label exited with an error; retrying in 30s" | tee -a "$LOG"
    sleep 30
  done
  echo "!!! $label failed 3 times; moving on" | tee -a "$LOG"
  return 1
}

echo "Logging to $LOG"
start=$(date +%s)
failed=0
run_step "Report covers (replace 100)" --report-covers || failed=1
run_step "Report cover pool (1,416)"   --report-pool  || failed=1
run_step "Compliance covers (842 topics × 6)"       || failed=1
"$PY" "$SCRIPT" --list-only >>"$LOG" 2>&1   # in case the last step died before writing the list

mins=$(( ($(date +%s) - start) / 60 ))
echo "=== $(date '+%F %T') finished in ${mins} min (failures: $failed). Log: $LOG" | tee -a "$LOG"
for d in acpwb/static/img/report-covers acpwb/static/img/report-covers/pool acpwb/static/img/compliance-covers; do
  printf '  %-45s %5s images  %s\n' "$d" "$(ls "$d"/*.webp 2>/dev/null | wc -l | tr -d ' ')" "$(du -sh "$d" 2>/dev/null | cut -f1)"
done
exit $failed
