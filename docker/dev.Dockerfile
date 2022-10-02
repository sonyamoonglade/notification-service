FROM golang:1.18

WORKDIR /app

RUN apt update -y && \
    apt upgrade -y && \
    apt install -y git && \
    go install github.com/githubnemo/CompileDaemon@latest

COPY . ./app

ENV APP_NAME=notification-service

ENTRYPOINT CompileDaemon -polling -build="go build -o ./bin/${APP_NAME} ./cmd/main.go" -command="./bin/${APP_NAME} --debug --strict=false"
