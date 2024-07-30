FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download -x
RUN go mod verify

COPY . .

RUN go build -v -o ./app ./cmd


FROM alpine
COPY --from=builder /app/app .

CMD ["./app"]
