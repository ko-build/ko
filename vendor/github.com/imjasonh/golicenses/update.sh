#!/usr/bin/env bash

set -euxo pipefail

time cat query.txt | bq query \
  --quiet \
  --nouse_legacy_sql \
  --format=csv \
  --headless=true \
  --max_rows=10000000 > licenses.csv

head licenses.csv
wc -l licenses.csv

gzip -f licenses.csv

go build -o out ./cmd/golicenses-example
ls -lh out licenses.*

time ./out github.com/google/go-containerregistry | grep Apache-2.0
rm out
