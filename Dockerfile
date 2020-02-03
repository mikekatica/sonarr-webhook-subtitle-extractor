FROM golang:1.13.7-buster

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT [ "sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
