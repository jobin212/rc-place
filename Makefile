GO           ?= go
BIN           = rc-place
SRC           = $(shell find . -type f -name '*.go')
.DEFAULT_GOAL = build

build: $(BIN)

run: $(BIN)
	./$<

$(BIN): $(SRC) go.mod go.sum home.html
	$(GO) build

clean:
	$(RM) $(BIN)

.PHONY: test
test:
	$(GO) test ./...

.PHONY: redis
redis:
	docker run --name rc-place-redis -d -p 6379:6379 redis
