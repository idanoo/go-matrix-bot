FROM golang:1.21.0

ARG DEBIAN_FRONTEND="noninteractive"

# linuxserver.io ftw <3
RUN apt-get update && apt-get install -y \
    apt-utils \
    locales && \
  echo "**** install packages ****" && \
  apt-get install -y \
    curl \
    gnupg \
    jq \
    tzdata \
    libolm-dev && \
  echo "**** generate locale ****" && \
  locale-gen en_US.UTF-8 && \
  echo "**** create abc user and make our folders ****" && \
  useradd -u 911 -U -d /config -s /bin/false abc && \
  usermod -G users abc && \
  mkdir -p \
    /app \
    /config && \
  echo "**** cleanup ****" && \
  apt-get autoremove && \
  apt-get clean && \
  rm -rf \
    /tmp/* \
    /var/lib/apt/lists/* \
    /var/tmp/* \
    /var/log/*

# Copy data across
COPY src /src
COPY migrations /app/migrations

# Copy run script
COPY run.sh /app/run.sh
RUN chmod +x /app/run.sh

# Build App
WORKDIR /src
RUN go build -o /app/main cmd/gomatrixbot/main.go 
RUN rm -fr /src

WORKDIR /app
CMD /app/run.sh
