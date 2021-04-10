FROM golang:1.16.3-buster as builder

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get && go build

FROM debian:buster

#Install mkvextract
RUN apt-get update && apt-get install -qy wget gnupg2
RUN echo 'deb https://mkvtoolnix.download/debian/ buster main' > /etc/apt/sources.list.d/mkvtoolnix.download.list && \
  wget -q -O - https://mkvtoolnix.download/gpg-pub-moritzbunkus.txt | apt-key add - && \
  apt-get update && apt-get -y install mkvtoolnix

WORKDIR /
COPY --from=builder /go/src/sonarr-webhook-subtitle-extractor/sonarr-webhook-subtitle-extractor /

ENTRYPOINT [ "/sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
