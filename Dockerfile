FROM alpine:latest

ADD admission-webhook /admission-webhook
ENTRYPOINT ["./admission-webhook"]