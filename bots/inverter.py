#!/usr/bin/python3
import os
import time
import requests
import random

PROD_URL = "https://rc-place.fly.dev/tile"
LOCAL_URL = "http://localhost:8080/tile"

# Wait period between API calls in seconds
TIMEOUT = .012


def get_tile(x_pos, y_pos):
    params = {"x": x_pos, "y": y_pos}
    headers = {'Authorization': 'Bearer ' + os.getenv("RC_PLACE_TOKEN"),
               'Content-Type': "application/json"}

    resp = requests.get(PROD_URL, params=params, headers=headers)

    if resp.status_code == 200:
        data = resp.json()
        return data["color"]
    else:
        print(
            f"Tile not pulled at ({x_pos},{y_pos}) | "
            f"Status code: {resp.status_code} | "
            f"Message: {resp.text}")
        return None


def swap_color(color):
    colors = {
        "black": "white",
        "white": "black",
        "green": "red",
        "red": "green",
        "orange": "blue",
        "blue": "orange",
        "yellow": "purple",
        "purple": "yellow",
        "forest": "pink",
        "lime": "hot-pink",
        "cornflowerblue": "burnt-orange",
        "sky": "cyan",
        "cyan": "sky",
        "burnt-orange": "cornflowerblue",
        "hot-pink": "lime",
        "pink": "forest",
    }

    return colors[color]


def set_tile(x_pos, y_pos, color):
    data = {"x": x_pos, "y": y_pos, "color": color}
    headers = {'Authorization': 'Bearer ' + os.getenv("RC_PLACE_TOKEN"),
               'Content-Type': "application/json"}

    resp = requests.post(PROD_URL, json=data, headers=headers)

    if resp.status_code == 200:
        print(f"Tile placed at ({x_pos},{y_pos})")
    else:
        print(
            f"Tile not placed at ({x_pos},{y_pos}) | "
            f"Status code: {resp.status_code} | "
            f"Message: {resp.text}")


def main():
    size_y = random.randrange(20)
    size_x = random.randrange(20)

    start_x = random.randrange(100-size_x)
    start_y = random.randrange(100-size_y)

    for y_pos in range(start_y, start_y+size_y):
        for x_pos in range(start_x, start_x+size_x):
            color = get_tile(x_pos, y_pos)
            invert = swap_color(color)
            set_tile(x_pos, y_pos, invert)
            time.sleep(TIMEOUT)


if __name__ == "__main__":
    main()
