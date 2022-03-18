#!/bin/bash

TARGET='https://rc-place.fly.dev'
# TARGET='http://localhost:8080'
TIMEOUT='1s'

tiles=(
    "black black black black black black black black black black black black"
    "black white white white white white white white white white white black"
    "black white black black black black black black black black white black"
    "black white lime  black lime  black lime  black black black white black"
    "black white black black black black black black black black white black"
    "black white black lime  lime  black lime  lime  black black white black"
    "black white black black black black black black black black white black"
    "black white black black black black black black black black white black"
    "black white white white white white white white white white white black"
    "black black black black black black black black black black black black"
    "skip  skip  skip  skip  black black black black skip  skip  skip  skip "
    "skip  black black black black black black black black black black skip "
    "black black black white black white black white black white black black"
    "black black white black white black white black white black black black"
    "black black black black black black black black black black black black"
)

function set-tile {
    curl "$TARGET/tile" \
        -d '{"x": '"$1"', "y": '"$2"', "color": "'"$3"'"}' \
        -H "Authorization: Bearer $TOKEN"
}

function set-tile-sleep {
    set-tile "$@"
    sleep "$TIMEOUT"
}

y=${2:-44}
for row in "${tiles[@]}"; do
    x=${1:-42}
    for tile in $row; do
        x=$(( x + 1 ))
        if [[ "$tile" == "skip" ]]; then
            continue
        fi
        set-tile-sleep "$x" "$y" "$tile"
    done
    y=$(( y + 1 ))
done
