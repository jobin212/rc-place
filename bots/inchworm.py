#!/usr/bin/python3

from threading import local
import requests
import os
import time

prod_url = "https://rc-place.fly.dev/tile"
local_url = "http://localhost:8080/tile"

# Wait period between API calls in seconds
timeout = 1.1

def set_tile(x, y, color):
    data = { "x": x, "y": y, "color": color}
    headers = { 'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN"), 'Content-Type': "application/json"}
    
    resp = requests.post(prod_url, json=data, headers=headers)
    
    if resp.status_code == 200:
        print("Tile placed at (%s,%s)" % (x, y))
    else:
        print("Tile not placed at (%s,%s) | Status code: %d | Message: %s" % (x, y, resp.status_code, resp.text))

def main():
    # create inchworm
    headX, headY, inchworm_length, inchworm_color, default_color = 75, 50, 5, "lime", "cornflowerblue"
    for y in range(inchworm_length):
        set_tile(headX, headY + y, inchworm_color)
        time.sleep(timeout)

    # move inchworm
    while True:
        set_tile(headX, headY + 1, inchworm_color)
        time.sleep(timeout)
        set_tile(headX, (headY-inchworm_length) % 100, default_color)
        time.sleep(timeout)
        headY = (headY + 1) % 100

if __name__ == "__main__":
    main()
