FROM golang:1.23-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/pos-api ./services/pos-api/cmd/api/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/pos-api /app/pos-api
EXPOSE 3000
ENTRYPOINT ["/app/pos-api"]
