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
RECURSE_URL = "https://www.recurse.com/api/v1/profiles/jobin212@gmail.com"

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
        # img_resized_intermediate = img_resized.convert(mode='P', colors=16)
        # img_resized_dithered = img_resized_intermediate.convert(mode='P', colors=16, dither=1, palette=Image.ADAPTIVE)
        img_resized_16 = img_resized.convert('P', palette=Image.ADAPTIVE, colors=16)
        img_resized_16.save("img1.png")
        pix = img_resized_16.convert('RGB').load()
        return pix


def get_color_from_rgb(rgb):
    hex = '#%02x%02x%02x' % rgb
    return hex


def main():
    image_url = get_profile_image()
    img_arr = get_image_array_from_url(image_url)
    color_set = set()
    for x in range(25):
        for y in range(25):
            color = get_color_from_rgb(img_arr[x, y])
            color_set.add(color)
            set_tile(y, x, "red")
            time.sleep(timeout)
    print(color_set)


if __name__ == "__main__":
    main()
