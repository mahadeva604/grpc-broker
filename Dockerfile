FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /grpc-broker ./cmd/server/main.go

FROM alpine

WORKDIR /

COPY --from=builder /grpc-broker /grpc-broker

EXPOSE 8086

CMD ["./grpc-broker"]