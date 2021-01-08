FROM golang:1.15.6-alpine3.12

RUN mkdir -p /src
WORKDIR /src
COPY . .
RUN go build -o app .
EXPOSE 8000
ENTRYPOINT ["./app"]
