# rc-place
[![Go Report Card](https://goreportcard.com/badge/github.com/jobin212/rc-place)](https://goreportcard.com/report/github.com/jobin212/rc-place) [![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0) ![Coverage: O](https://img.shields.io/badge/coverage-200%25-red)



A place for [Recursers](https://www.recurse.com) to color pixels, inspired by
[r/place](https://www.reddit.com/r/place/). For architecture details, see [How We Built r/Place
](https://www.redditinc.com/blog/how-we-built-rplace).

![RC Place Image from 2022-03-21](docs/rc-place-2022-03-21.png)

## Build and Run
Create an OAuth application at [https://www.recurse.com/settings/apps](https://www.recurse.com/settings/apps) with proper redirect URI (http://localhost:8080/auth for local run).
**Make sure to set your app's ID, Secret, and Redirect URI in your environmental variables** (see [.env.example](.env.example)). You can optionally set your own redis host and password.

```shell
# Load your environmental variables after setting them
🎨 source .env.example

# Run Redis via docker container
🎨 docker run --name rc-place-redis -d -p 6379:6379 redis

# Run rc-place app
🎨 make run
```
🎉 rc-place should now be running at [http://localhost:8080](http://localhost:8080)

## Other tools

```shell
# Use Redis docker
🎨 docker exec -it rc-place-redis redis-cli

# Reset board
🎨 del $REDIS_BOARD_KEY 

# Get board at offset (x + boardSize*y)
🎨 bitfield $REDIS_BOARD_KEY GET u4 #$OFFSET
```

## Deploy
```shell
🎨 fly deploy
```

## Rest API

**Update tile**
----
Update the color of a tile located at column x, row y.
* **URL:** /tile
* **Method:** `POST`
* **Data Params:**
```json
{
    `x` int,
    `y` int,
    `color` string
}
```
Valid colors: `black`, `forest`, `green`, `lime`,          `blue`, `cornflowerblue`, `sky`, `cyan`, `red`, `burnt-orange`, `orange`, `yellow`, `purple`, `hot-pink`, `pink`, `white`.

* **Success Response:** 200
* **Error Response**
  * **Code** 400 Bad Request <br />
    * Invalid json body, make sure you're using the right types and your body is encoded correctly.
  * **Code** 401 Unauthorized <br />
    * Make sure you have a valid personal access token in your authorization header.
  * **Code** 422 Too Early <br />
    * There's a time limit for sending requets, make sure to wait one second between requests.
  * **Code** 500 Internal Server Error <br />
    * You may have found a bug! You're encourage to file an issue with the steps to reproduce.

* **Sample Call**
```shell
🎨 curl -X POST http://localhost:8080/tile -H "Content-Type: application/json" -d '{"x": 3, "y": 3, "color": "red"}' -H "Authorization: Bearer $PERSONAL_ACCESS_TOKEN"
```

**Get Tiles**
----
Get all tiles.
* **URL:** /tiles
* **Method:** `GET`
* **Success Response:** 200
```json 
{
  // 2d array of integers representing board state
}
```
* **Error Response**
  * **Code** 401 Unauthorized <br />
    * Make sure you have a valid personal access token in your authorization header.
  * **Code** 500 Internal Server Error <br />
    * You may have found a bug! You're encourage to file an issue with the steps to reproduce.

* **Sample Call**
```shell
🎨 curl http://localhost:8080/tiles -H "Authorization: Bearer $PERSONAL_ACCESS_TOKEN"
```





## TODO
- Follow up on other architecture decisions detailed in [How We Built r/place](https://www.redditinc.com/blog/how-we-built-rplace), e.g. Redis
- Add API to load entire board state, update frontend to follow
- Make board big (1000 x 1000) and full screen
- Add user information to tiles -- who placed what tile
- Add github actions
- Generate timelapse (use a timeseries db -- influxdb?)
- Formal websockets message
- Give web client a message or timer until they can place again -- e.g. you have 0.4 s until you can update
- Load testing
- Token server?
- Queue up changes and apply them all at midnight?
- SSH clients
