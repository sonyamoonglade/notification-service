FROM golang:1.18 as builder

WORKDIR /app/notification

COPY . /app/notification

RUN mkdir bin && \
    CGO_ENABLED=0 GOOS=linux go build -o ./bin/app ./cmd/main.go


FROM alpine:latest as prod

WORKDIR /app/notification

RUN mkdir bin

COPY --from=builder /app/notification/bin ./bin
COPY --from=builder /app/notification/prod.config.yaml .

CMD ["sh","-c","bin/app"]
