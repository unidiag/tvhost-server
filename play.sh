#!/bin/bash

FILES=("1.ts" "2.ts" "3.ts")
play_file() {
  local file=$1
  tsplay "$file" -i 127.0.0.1 -udp 239.1.1.1:1234
}
for file in "${FILES[@]}"; do
  play_file "$file"
done