APP_NAME=gwatch
APP_BINARY=./bin/${APP_NAME}.exe
APP_BIN_PATH=./bin/

all: build run

run:
	@${APP_BINARY}

build:
	@go build -o ${APP_BIN_PATH} ./cmd/${APP_NAME}