FROM golang:1.14.4

RUN mkdir -p src/go-gitlab-webhook
WORKDIR src/go-gitlab-webhook
COPY . .
RUN go build .
EXPOSE 8000
ENTRYPOINT ["./go-gitlab-webhook"]
