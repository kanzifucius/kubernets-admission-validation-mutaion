FROM alpine:latest

ADD bin/admissionwebhook_unix  /admission-webhook
ENTRYPOINT ["./admission-webhook"]