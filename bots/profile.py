#!/usr/bin/python3

from email.policy import default
from threading import local
import requests
import os
import time
import sys
from PIL import Image
from io import BytesIO

# url = "https://rc-place.fly.dev/tile"
url = "http://localhost:8080"
default_color = "cornflowerblue"

# the url to get the profile of the owner of the auth token
recurse_url = "https://www.recurse.com/api/v1/profiles/me"

# Wait period between API calls in seconds
timeout = 0.011

hexToName = {
    "#000000": "black",
    "#005500": "forest",
    "#00ab00": "green",
    "#00ff00": "lime",
    "#0000ff": "blue",
    "#6495ed": "cornflowerblue",
    "#00abff": "sky",
    "#00ffff": "cyan",
    "#ff0000": "red",
    "#ff5500": "burnt-orange",
    "#ffab00": "orange",
    "#ffff00": "yellow",
    "#6a0dad": "purple",
    "#ff55ff": "hot-pink",
    "#ffabff": "pink",
    "#ffffff": "white",
}


def set_tile(x, y, color):
    data = {"x": x, "y": y, "color": color}
    headers = {'Authorization': 'Bearer ' +
               os.getenv("PERSONAL_ACCESS_TOKEN"), 'Content-Type': "application/json"}
    formatted_url = "{0}/tile".format(url)

    resp = requests.post(formatted_url, json=data, headers=headers)

    if resp.status_code == 200:
        print("Tile placed at (%s,%s)" % (x, y))
    else:
        print("Tile not placed at (%s,%s) | Status code: %d | Message: %s" %
              (x, y, resp.status_code, resp.text))


def get_profile_image():
    headers = {'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN")}
    resp = requests.get(RECURSE_URL, headers=headers)
    if resp.status_code == 200:
        resp_body = resp.json()
        return resp_body['image_path']
    else:
        return 'no_image'


def get_image_array_from_url(image_url):
    resp = requests.get(image_url)
    with Image.open(BytesIO(resp.content)) as img:
        img_resized = img.resize((25, 25))

        pal_image = Image.new("P", (1, 1))
        pal_image.putpalette((
            0, 0, 0,  # 000000
            0, 85, 0,  # 005500
            0, 171, 0,  # 00ab00
            0, 255, 0,  # 00ff00
            0, 0, 255,  # 0000ff
            100, 149, 237,  # 6495ed
            0, 171, 255,  # 00abff
            0, 255, 255,  # 00ffff
            255, 0, 0,  # ff0000
            255, 85, 0,  # ff5500
            255, 171, 0,  # ffab00
            255, 255, 0,  # ffff00
            106, 13, 173,  # 6a0dad
            255, 85, 255,  # ff55ff
            255, 171, 255,  # ffabff
            255, 255, 255,  # ffffff
        )
            + (0, 0, 0) * 240
        )

        img_rs_q = img_resized.convert("RGB").quantize(palette=pal_image)
        return img_rs_q.convert("RGB").load()


def get_color_from_rgb(rgb):
    hex = '#%02x%02x%02x' % rgb
    return hexToName[hex]


def main(args):
    offsetX, offsetY = 0, 0
    try:
        offsetX, offsetY = int(args[0]), int(args[1])
    except:
        print("Invalid x, y values given. Starting inchworm at %d,%d" %
              (offsetX, offsetY))

    image_url = get_profile_image()
    img_arr = get_image_array_from_url(image_url)
    for y in range(25):
        for x in range(25):
            color = get_color_from_rgb(img_arr[x, y])
            set_tile(offsetX + x, offsetY + y, color)
            time.sleep(timeout)


if __name__ == "__main__":
    main(sys.argv[1:])
