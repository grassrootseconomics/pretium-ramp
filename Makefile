BIN := pretium-ramp
BUILD_CONF := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null)
DEBUG := DEV=true

.PHONY: build run clean

clean:
	rm ${BIN}

build:
	${BUILD_CONF} go build -ldflags="-X main.build=${BUILD_COMMIT} -s -w" -o build/${BIN} cmd/main.go

run:
	${BUILD_CONF} ${DEBUG} go run cmd/main.go