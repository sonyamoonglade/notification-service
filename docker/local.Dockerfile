FROM golang:1.18

WORKDIR /app/notification

COPY . /app/notification

RUN mkdir bin && \
    CGO_ENABLED=0 GOOS=linux go build -o ./bin/app ./cmd/main.go

CMD ["sh", "-c","bin/app"]
