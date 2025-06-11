FROM golang:1.23-alpine as builder
WORKDIR /src
COPY . .
RUN apk add --no-cache git
RUN go build -o /server ./cmd/server

FROM alpine
COPY --from=builder /server /app/server
CMD ["/app/server"]
