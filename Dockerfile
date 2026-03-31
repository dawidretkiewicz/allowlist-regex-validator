FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/server /app/server
COPY --from=builder /src/web/static /app/web/static

USER app
EXPOSE 8080
ENV ADDR=:8080

ENTRYPOINT ["/app/server"]
