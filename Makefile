test:
	go test ./... -count=1 -parallel=10 -v

lint:
	golangci-lint run
