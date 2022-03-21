# rc-place
A place for [Recursers](https://www.recurse.com) to color pixels, inspired by
[r/place](https://www.reddit.com/r/place/). For architecture details, see [How We Built r/Place
](https://www.redditinc.com/blog/how-we-built-rplace).

![RC Place Image from 2022-03-21](docs/rc-place-2022-03-21.png)

## Build and Run
- Create an OAuth application at [https://www.recurse.com/settings/apps](https://www.recurse.com/settings/apps) with proper redirect URI (http://localhost:8080/auth for local run). Make sure to set your app's ID, Secret, and Redirect URI in your environmental vairables.
- Run docker Redis.
```bash
docker run redis
```
```bash
make run
```

## Diagnose
- Access local 


## Deploy
```bash
fly deploy
```

## TODO
- Follow up on other architecture decisions detailed in [How We Built r/place](https://www.redditinc.com/blog/how-we-built-rplace), e.g. Redis
- Add API to load entire board state, update frontend to follow
- Make board big (1000 x 1000) and full screen
- Look into performance we could get on frontend or backend
- Add user information to tiles -- who placed what tile
- Bot example
- Refactor main.go, get a good go score
- Add github actions, make it so @miccah can deploy
- Generate timelapse (use a timeseries db -- influxdb?)
- Formal websockets message
- Give web client a message or timer until they can place again -- e.g. you have 0.4 s until you can update
- Load testing
- Token server?
- Explore RC cluster?
- Queue up changes and apply them all at midnight?
