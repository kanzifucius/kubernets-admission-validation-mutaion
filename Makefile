# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=admissionwebhook
BINARY_UNIX=$(BINARY_NAME)_unix


all:  clean test build docker-build
build:
	dep ensure -v
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v  ./cmd/admissionwebhook/
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	$(GOGET) github.com/markbates/goth
	$(GOGET) github.com/markbates/pop

docker-build:
		docker build --no-cache -t vod-docker-ms.artifactory.vodacom.co.za/admission-webhook:v1 .
		docker push vod-docker-ms.artifactory.vodacom.co.za/admission-webhook:v1

