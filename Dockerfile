FROM golang:1.22 as builder

RUN mkdir -p /app

WORKDIR /app

COPY . /app

RUN go mod tidy

RUN CGO_ENABLED=0 go build -o fivem-server-analytics ./

RUN chmod +x /app/fivem-server-analytics

FROM alpine:latest

RUN mkdir -p /app

ENV TZ=Europe/Amsterdam

COPY --from=builder /app/fivem-server-analytics /app
COPY --from=builder /app/.env.production /app/.env

CMD ["/app/fivem-server-analytics"]