FROM golang:1.13.7-buster as builder

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get github.com/mattn/gom
RUN go get github.com/go-playground/validator
RUN gom install && gom build

FROM debian:buster

#Install mkvextract
RUN echo 'deb https://mkvtoolnix.download/debian/ buster main' > /etc/apt/sources.list.d/mkvtoolnix.download.list && \
  wget -q -O - https://mkvtoolnix.download/gpg-pub-moritzbunkus.txt | apt-key add - && \
  apt-get update && apt-get -y install mkvtoolnix

WORKDIR /
COPY --from=builder /go/src/sonarr-webhook-subtitle-extractor/sonarr-webhook-subtitle-extractor

ENTRYPOINT [ "/sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
