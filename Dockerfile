FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o vizen-api ./cmd/api

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/vizen-api .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD [ "./vizen-api" ]
