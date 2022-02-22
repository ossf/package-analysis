#!/bin/sh -x
# Script to bulk request backfills in parallel.

NUM_WORKERS=128

if [ $# -lt 2 ]; then
  echo "Usage: $0 <path to package list> <ecosystem>"
  exit 1
fi

cat $1 | xargs -I {} -P $NUM_WORKERS -n 1 python3 analysis_runner.py -a -n {} $2
