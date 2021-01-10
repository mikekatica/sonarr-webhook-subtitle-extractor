FROM golang:1.13.7-buster as builder

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get github.com/mattn/gom
RUN gom install && gom build

ENTRYPOINT [ "sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
