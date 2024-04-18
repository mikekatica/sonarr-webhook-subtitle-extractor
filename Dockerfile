FROM golang:1.22.2-bookworm as builder

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get && go build

FROM linuxserver/ffmpeg:7.0-cli-ls132

WORKDIR /
COPY --from=builder /go/src/sonarr-webhook-subtitle-extractor/sonarr-webhook-subtitle-extractor /
COPY public/ /public

ENTRYPOINT [ "/sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
