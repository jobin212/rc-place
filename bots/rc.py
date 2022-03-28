#!/usr/bin/python3

import requests
import os
import time

tiles=(
    "black black black black black black black black black black black black",
    "black white white white white white white white white white white black",
    "black white black black black black black black black black white black",
    "black white lime  black lime  black lime  black black black white black",
    "black white black black black black black black black black white black",
    "black white black lime  lime  black lime  lime  black black white black",
    "black white black black black black black black black black white black",
    "black white black black black black black black black black white black",
    "black white white white white white white white white white white black",
    "black black black black black black black black black black black black",
    "skip  skip  skip  skip  black black black black skip  skip  skip  skip ",
    "skip  black black black black black black black black black black skip ",
    "black black black white black white black white black white black black",
    "black black white black white black white black white black black black",
    "black black black black black black black black black black black black"
)

prod_url = "https://rc-place.fly.dev/tile"
local_url = "http://localhost:8080/tile"

def set_tile(x, y, color):
    data = { "x": x, "y": y, "color": color}
    headers = { 'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN"), 'Content-Type': "application/json"}
    resp = requests.post(prod_url, json=data, headers=headers)
    if resp.status_code == 200:
        print("Tile placed at (%s,%s)" % (x, y))
    else:
        print("Tile not placed at (%s,%s) | Status code: %d | Message: %s" % (x, y, resp.status_code, resp.text))

def main():
    offsetX, offsetY = 5, 5
    for y in range(len(tiles)):
        colors = tiles[y].split()
        for x, color in enumerate(colors):
            if color != "skip":
                set_tile(offsetX + x, offsetY + y, color)
                time.sleep(1.1)

if __name__ == "__main__":
    main()
