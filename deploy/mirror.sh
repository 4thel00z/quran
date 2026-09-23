#!/usr/bin/env bash
# Mirror ayah MP3s into the quran.host web root. Reads "url path" lines on stdin
# (from `quran mirror`), skips files already present, 8 downloads in parallel.
set -euo pipefail
root="${1:-$HOME/private/quran.host/www}"
mkdir -p "$root"
cd "$root"
xargs -P 8 -L 1 sh -c '
  [ -s "$1" ] && exit 0
  mkdir -p "$(dirname "$1")"
  curl -sfL --retry 5 --retry-delay 2 -o "$1.part" "$0" && mv "$1.part" "$1" || echo "failed $0" >&2
'
