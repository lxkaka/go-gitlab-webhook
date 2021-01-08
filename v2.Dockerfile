FROM golang:1.15.6-alpine3.12 as builder

RUN mkdir -p /src
WORKDIR /src
COPY . .
RUN go build -o app .

FROM alpine:3.12

COPY --from=builder /src/ .
EXPOSE 8000
ENTRYPOINT ["./app"]