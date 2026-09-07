FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -a -o /out/app ./

FROM alpine:3.20
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /out/app /app/app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
