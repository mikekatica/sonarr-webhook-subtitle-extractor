FROM golang:1.16.3-buster as builder

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get && go build

FROM jrottenberg/ffmpeg:4-ubuntu

WORKDIR /
COPY --from=builder /go/src/sonarr-webhook-subtitle-extractor/sonarr-webhook-subtitle-extractor /

ENTRYPOINT [ "/sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
