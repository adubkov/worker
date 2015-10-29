all: 
	@mkdir -p bin
	@GOOS=darwin GOARCH=amd64 go build -o ./bin/worker.darwin.amd64 *.go
	@GOOS=linux GOARC=amd64 go build -o ./bin/worker.linux.amd64 *.go

.PHONY: all
