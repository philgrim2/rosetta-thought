# Copyright 2020 Coinbase, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build thoughtd
FROM ubuntu:20.04 as thoughtd-builder

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

# Source: https://github.com/thoughtnetwork/thought/blob/master/doc/build-unix.md#ubuntu--debian
ARG DEBIAN_FRONTEND=noninteractive
ENV TZ Etc/UTC
RUN apt-get update && apt-get install -y curl wget cmake default-jdk make gcc g++ autoconf autotools-dev bsdmainutils build-essential git libboost-all-dev \
  libcurl4-openssl-dev libdb++-dev libevent-dev libssl-dev libtool pkg-config python python3-pip libzmq3-dev wget

# VERSION: Thought Core 0.18.3
RUN git clone https://github.com/thoughtnetwork/thought \
  && cd thought 

RUN cd thought \
  && ./configure-static.sh --disable-tests --without-miniupnpc --without-gui --with-incompatible-bdb --disable-hardening --disable-zmq --disable-bench --disable-wallet \
  && make

RUN mv thought/src/thoughtd /app/thoughtd \
  && mv thought/src/thought-cli /app/thought-cli \
  && rm -rf thought

# Build Rosetta Server Components
FROM ubuntu:20.04 as rosetta-builder

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

RUN apt-get update && apt-get install -y curl make gcc g++
# Install Golang 1.20
ENV GOLANG_VERSION 1.20
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 5a9ebcc65c1cce56e0d2dc616aff4c4cedcfbda8cc6f0288cc08cda3b18dcbf1

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Use native remote build context to build in any directory
COPY . src 
RUN cd src \
  && go build \
  && cd .. \
  && mv src/rosetta-thought /app/rosetta-thought \
  && mv src/assets/* /app \
  && rm -rf src 

## Build Final Image
FROM ubuntu:20.04

RUN apt-get update && \
  apt-get install --no-install-recommends -y wget libevent-dev libboost-system-dev libboost-filesystem-dev libboost-test-dev libboost-thread-dev && \
  apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app \
  && mkdir -p /data/thoughtd/testnet3 \
  && wget --no-check-certificate -O /data/thoughtd/testnet3/testchain.tar.gz https://idea-01.insufficient-light.com/data/testchain.tar.gz \
  && tar -C /data/thoughtd/testnet3 -xzf /data/thoughtd/testnet3/testchain.tar.gz  \
  && rm  /data/thoughtd/testnet3/testchain.tar.gz \
  && chown -R nobody:nogroup /data

WORKDIR /app

# Copy binary from thoughtd-builder
COPY --from=thoughtd-builder /app/thoughtd /app/thoughtd
COPY --from=thoughtd-builder /app/thought-cli /app/thought-cli

# Copy binary from rosetta-builder
COPY --from=rosetta-builder /app/* /app/

# Set permissions for everything added to /app
RUN chmod -R 755 /app/*

CMD ["/app/rosetta-thought"]
