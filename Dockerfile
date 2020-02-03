FROM golang:1.13.7-buster

#Install mkvextract
RUN echo 'deb https://mkvtoolnix.download/debian/ buster main' > /etc/apt/sources.list.d/mkvtoolnix.download.list && \
  wget -q -O - https://mkvtoolnix.download/gpg-pub-moritzbunkus.txt | apt-key add - && \
  apt-get update && apt-get -y install mkvtoolnix

WORKDIR /go/src/sonarr-webhook-subtitle-extractor
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT [ "sonarr-webhook-subtitle-extractor", "--alsologtostderr" ]
CMD []
