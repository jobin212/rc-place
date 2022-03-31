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

def set_tile(x, y, color):
    data = { "x": x, "y": y, "color": color}
    headers = { 'Authorization': 'Bearer ' + os.getenv("PERSONAL_ACCESS_TOKEN"), 'Content-Type': "application/json"}
    formatted_url = "{0}/tile".format(url)
    
    resp = requests.post(formatted_url, json=data, headers=headers)
    
    if resp.status_code == 200:
        print("Tile placed at (%s,%s)" % (x, y))
    else:
        print("Tile not placed at (%s,%s) | Status code: %d | Message: %s" % (x, y, resp.status_code, resp.text))

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
        pix = img.load()
        return pix

def get_color_from_rgb(rgb):
    r, g, b = rgb
    return "red"

def main():
    image_url = get_profile_image()
    img_arr = get_image_array_from_url(image_url)
    for x in range(25):
        for y in range(25):
            color = get_color_from_rgb(img_arr[x, y])
            set_tile(y, x, color)
            time.sleep(timeout)

if __name__ == "__main__":
    main()
