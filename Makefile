APP_NAME=gwatch
APP_PATH=./cmd/${APP_NAME}
APP_BIN_PATH=./bin

all: build run

run:
	go run ${APP_PATH}

build:
	GOOS=linux GOARCH=amd64 go build -o ${APP_BIN_PATH}/linux/${APP_NAME} ${APP_PATH}
	GOOS=darwin GOARCH=amd64 go build -o ${APP_BIN_PATH}/darwin/${APP_NAME} ${APP_PATH}
	GOOS=windows GOARCH=amd64 go build -o ${APP_BIN_PATH}/windows/${APP_NAME}.exe ${APP_PATH}