FROM golang:1.24.1-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -extldflags=-static" -tags sqlite_omit_load_extension -o /app/server ./cmd/app/main.go

FROM alpine:3.19

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /app/server /app/server

COPY config.yaml /app/config.yaml

COPY migrations ./migrations

RUN mkdir /app/data # && chown nonroot:nonroot /app/data

EXPOSE 8080

ENTRYPOINT ["/app/server"]