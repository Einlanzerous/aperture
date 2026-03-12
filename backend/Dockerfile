FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod ./
COPY go.sum* ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o aperture ./cmd/server

FROM scratch
COPY --from=builder /app/aperture /aperture
EXPOSE 8080
ENTRYPOINT ["/aperture"]
