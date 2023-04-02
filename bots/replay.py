#!/usr/bin/python3

import requests
import os
import time
import sys
from typing import Tuple, List


color_to_name = {
    0:  'black',
    1:  'forest',
    2:  'green',
    3:  'lime',
    4:  'blue',
    5:  'cornflowerblue',
    6:  'sky',
    7:  'cyan',
    8:  'red',
    9:  'burnt-orange',
    10: 'orange',
    11: 'yellow',
    12: 'purple',
    13: 'hot-pink',
    14: 'pink',
    15: 'white',
}

# todo(joseph-pr) set updateLimitInMs to 0s
# url = 'https://rc-place.fly.dev'
url = 'http://localhost:8080'


def set_tile(x, y, color_id):
    data = {'x': x, 'y': y, 'color': color_to_name[color_id]}
    headers = {
        'Authorization': 'Bearer ' + os.getenv('PERSONAL_ACCESS_TOKEN'),
        'Content-Type': 'application/json',
    }
    formatted_url = f'{url}/tile'

    resp = requests.post(formatted_url, json=data, headers=headers)

    if resp.status_code == 200:
        print(f'Tile placed at ({x},{y})')
    else:
        print(f'Tile not placed at ({x},{y}) | Status code: {resp.status_code} | Message: {resp.text}')

def parse_aof_file_into_bot_commands(file_name: str) -> List[Tuple[int, int]]:
    commands = []
    with open(file_name, 'rb') as f:
        while True:
            line = f.readline()
            if len(line) == 0:
                break
            if line.strip() == b'SET':
                try:
                    f.readline()
                    f.readline()
                    f.readline()
                    raw_position = f.readline().decode()
                    f.readline()
                    raw_color = f.readline().decode()

                    position = int(raw_position[1:])
                    color = int(raw_color)

                    if not 0 <= position < 10000 or not 0 <= color < 16:
                        continue

                    commands.append((position, color))
                except Exception as e:
                    print(f"Continuing after exception {e}")
                    continue
    return commands


def main(args):
    commands = parse_aof_file_into_bot_commands(args[0])

    for pos_index, color_id in commands:
        pos_x, pos_y = pos_index % 100, pos_index // 100
        set_tile(pos_x, pos_y, color_id)
        time.sleep(0.001)


if __name__ == '__main__':
    main(sys.argv[1:])
