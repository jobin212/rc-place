#!/usr/bin/python3

from email.policy import default
from threading import local
import requests
import os
import time
import sys

# url = "https://rc-place.fly.dev/tile"
url = "http://localhost:8080"
default_color = "cornflowerblue"

# Wait period between API calls in seconds
timeout = 0.011

def set_tile(x, y, color):
    data = { "x": x, "y": y, "color": color}
    headers = { 'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN"), 'Content-Type': "application/json"}
    formatted_url = "{0}/tile".format(url)
    
    resp = requests.post(formatted_url, json=data, headers=headers)
    
    if resp.status_code == 200:
        print("Tile placed at (%s,%s)" % (x, y))
    else:
        print("Tile not placed at (%s,%s) | Status code: %d | Message: %s" % (x, y, resp.status_code, resp.text))

def get_tile(x, y):
    headers = { 'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN"), 'Content-Type': "application/json"}

    formatted_url = "{0}/{1}?x={2}&y={3}".format(url, "tile", x, y)
    resp = requests.get(formatted_url, headers=headers)

    if resp.status_code == 200:
        resp_body = resp.json()
        color = resp_body["color"]
        return color
    else:
        print("GET failed")
        return default_color

def get_tiles():
    headers = { 'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN")}
    formtted_url = "{0}/{1}".format(url, "tiles")
    resp = requests.get(formtted_url, headers=headers)

    if resp.status_code == 200:
        resp_body = resp.json()
        return resp_body["tiles"]
    else:
        print("GET failed")
        return None

def get_most_popular_color():
    tiles = get_tiles()

    color_count = {}

    if tiles is not None:
        for y in range(len(tiles)):
            for x in range(len(tiles[y])):
                color = tiles[x][y]
                if color in color_count:
                    color_count[color] += 1
                else :
                    color_count = 1

    print(color_count)

    return max(color_count, key=color_count.get)

def main(args):
    # grow longer only when we see the least popular color
    default_color = get_most_popular_color()

    print(default_color)

    # headX, headY = 76, 25
    # try:
    #     headX, headY = int(args[0]), int(args[1])
    # except:
    #     print("Invalid x, y values given. Starting inchworm at {0},{1}", headX, headY)

    # # create inchworm
    # inchworm_length, inchworm_color = 5, "burnt-orange"

    # for y in reversed(range(inchworm_length)):
    #     set_tile(headX, (headY - y)%100, inchworm_color)
    #     time.sleep(timeout)

    # # move inchworm
    # while True:
    #     color = get_tile(headX, headY + 1)

    #     set_tile(headX, headY + 1, inchworm_color)
    #     time.sleep(timeout)
    #     if color != default_color:
    #         # inchworm just ate a non-default block and can now grow
    #         inchworm_length += 1
    #     else:
    #         # move the tail of the inchworm
    #         set_tile(headX, (headY-inchworm_length) % 100, default_color)
    #         time.sleep(timeout)
    #     headY = (headY + 1) % 100

if __name__ == "__main__":
    main(sys.argv[1:])
