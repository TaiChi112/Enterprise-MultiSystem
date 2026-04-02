FROM golang:1.23-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/oms-api ./services/oms-api/cmd/api/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/oms-api /app/oms-api
EXPOSE 4003
ENTRYPOINT ["/app/oms-api"]
