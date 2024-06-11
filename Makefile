DOCKER_CONTAINER=fivem-server-analytics-websocket-server
DOCKER_IMAGE=fivem-server-analytics

build:
	@go build -o bin/fsa

run: build
	@./bin/fsa

test:
	@go test ./...


docker:
	@docker rm $(DOCKER_CONTAINER) || true
	@docker build -t $(DOCKER_IMAGE) .
	@docker run -d -p 5000:5000 --name $(DOCKER_CONTAINER) $(DOCKER_IMAGE)