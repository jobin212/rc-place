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
	docker run --rm --name rc-place-redis -d -p 6379:6379 redis

.PHONY: pg
pg:
	docker run --rm --name rc-place-pg -d \
		-e POSTGRES_DB=metadata -e POSTGRES_HOST_AUTH_METHOD=trust \
		-p 5432:5432 postgres

.PHONY: start-docker
start-docker: redis pg

.PHONY: stop-docker
stop-docker: stop-redis stop-pg

.PHONY: stop-redis
stop-redis:
	docker stop rc-place-redis || true

.PHONY: stop-redis
stop-pg:
	docker stop rc-place-pg || true
