
build:
	@go build -o bin/fsa

run: build
	@./bin/fsa

test:
	@go test ./...


docker:
	@docker build -t fivem-server-analytics .
	@docker run -d -p 5000:5000 fivem-server-analytics
