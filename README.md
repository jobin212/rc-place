# rc-place

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