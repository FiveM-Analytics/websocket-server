
build:
	@go build -o bin/fsa

run: build
	@./bin/fsa

test:
	@go test ./...