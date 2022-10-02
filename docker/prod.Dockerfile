FROM golang:1.18 as builder

WORKDIR /app/notification

COPY . /app/notification

RUN mkdir bin && \
    CGO_ENABLED=0 GOOS=linux go build -o ./bin/app ./cmd/main.go


FROM alpine:latest as prod

WORKDIR /app/notification

RUN mkdir bin && \
    mkdir logs && \
    touch logs/log.txt


COPY --from=builder /app/notification/bin ./bin
COPY --from=builder /app/notification/events.json .
COPY --from=builder /app/notification/templates.json .

CMD ["sh","-c","bin/app"]
