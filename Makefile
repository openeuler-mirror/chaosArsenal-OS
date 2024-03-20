build:
	go build -ldflags "-linkmode internal -extldflags -static"

clean:
	go clean ./...