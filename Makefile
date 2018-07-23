deps:
	@go get ./...

run:
	@go build -o bin/ava
	@./bin/ava
