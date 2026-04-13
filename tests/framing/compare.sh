#!/usr/bin/env bash
# compare.sh — Run A/B framing comparison tests
# Usage: ./compare.sh [runs_per_variant] [model]
#
# Runs each scenario pair (negative vs positive framing) multiple times
# and compares pass rates to measure the effect of instruction framing.
set -uo pipefail

RUNS=${1:-5}
MODEL=${2:-haiku}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
SCENARIO_DIR="$SCRIPT_DIR/scenarios"

cd "$TEST_DIR" || { echo "error: cannot cd to $TEST_DIR"; exit 1; }

SYS_PROMPT="Answer based ONLY on the convention text provided in the prompt. Do not use tools or read repository files."

pairs=("01:gcp-project" "02:docker-image" "03:health-check" "04:file-extension" "05:namespace")

echo "=== Instruction Framing A/B Test ==="
echo "Runs per variant: $RUNS | Model: $MODEL"
echo ""

declare -a neg_scores=()
declare -a pos_scores=()

for entry in "${pairs[@]}"; do
  IFS=: read -r pair name <<< "$entry"
  neg_pass=0
  pos_pass=0

  echo "--- Pair $pair: $name ---"

  for i in $(seq 1 "$RUNS"); do
    # Negative variant
    output=$(go run . --dir "$SCENARIO_DIR" --scenario "f${pair}a-*" \
      --model "$MODEL" --no-retry \
      --system-prompt "$SYS_PROMPT" --allowed-tools none 2>&1) || true
    if echo "$output" | grep -q " PASS "; then
      ((neg_pass++)) || true
    fi

    # Positive variant
    output=$(go run . --dir "$SCENARIO_DIR" --scenario "f${pair}b-*" \
      --model "$MODEL" --no-retry \
      --system-prompt "$SYS_PROMPT" --allowed-tools none 2>&1) || true
    if echo "$output" | grep -q " PASS "; then
      ((pos_pass++)) || true
    fi

    printf "  Run %d/%d  neg=%d  pos=%d\n" "$i" "$RUNS" "$neg_pass" "$pos_pass"
  done

  neg_scores+=("$neg_pass")
  pos_scores+=("$pos_pass")
  echo ""
done

# Summary table
echo "=== Results ==="
printf "%-20s  %-12s  %-12s  %s\n" "Topic" "Negative" "Positive" "Winner"
printf "%-20s  %-12s  %-12s  %s\n" "-----" "--------" "--------" "------"

total_neg=0
total_pos=0

for idx in "${!pairs[@]}"; do
  IFS=: read -r pair name <<< "${pairs[$idx]}"
  neg=${neg_scores[$idx]}
  pos=${pos_scores[$idx]}
  ((total_neg += neg)) || true
  ((total_pos += pos)) || true

  delta=$((pos - neg))
  if [ "$delta" -gt 0 ]; then
    winner="Positive (+$delta)"
  elif [ "$delta" -lt 0 ]; then
    winner="Negative (+$((-delta)))"
  else
    winner="Tie"
  fi

  printf "%-20s  %3d/%-8s  %3d/%-8s  %s\n" "$name" "$neg" "$RUNS" "$pos" "$RUNS" "$winner"
done

echo ""
total_runs=$((RUNS * ${#pairs[@]}))
total_delta=$((total_pos - total_neg))
if [ "$total_delta" -gt 0 ]; then
  total_winner="POSITIVE (+$total_delta)"
elif [ "$total_delta" -lt 0 ]; then
  total_winner="NEGATIVE (+$((-total_delta)))"
else
  total_winner="TIE"
fi

printf "%-20s  %3d/%-8s  %3d/%-8s  %s\n" "TOTAL" "$total_neg" "$total_runs" "$total_pos" "$total_runs" "$total_winner"
