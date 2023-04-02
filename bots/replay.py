#!/usr/bin/python3

import requests
import os
import time
import sys


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


def main(args):
    for line in open('commands.txt', 'r').readlines():
        pos_index, color_id = map(int, line.split(':'))
        pos_x, pos_y = pos_index % 100, pos_index // 100
        set_tile(pos_x, pos_y, color_id)
        time.sleep(0.001)


if __name__ == '__main__':
    main(sys.argv[1:])
