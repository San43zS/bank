FROM golang:1.23-alpine AS build

RUN apk add --no-cache ca-certificates git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/banking-platform ./cmd/app
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

FROM alpine:3.20 AS runner
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/banking-platform /app/banking-platform
EXPOSE 8080
CMD ["/app/banking-platform"]

FROM alpine:3.20 AS migrator
RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY migration /migrations
CMD ["sh", "-c", "goose -dir /migrations up"]

