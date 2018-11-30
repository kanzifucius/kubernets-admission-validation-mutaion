#!/bin/bash
dep ensure -v
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admission-webhook
docker build --no-cache -t vod-docker-ms.artifactory.vodacom.co.za/admission-webhook:v1 .
rm -rf admission-webhook
docker push vod-docker-ms.artifactory.vodacom.co.za/admission-webhook:v1