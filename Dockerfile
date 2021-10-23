FROM docker.io/library/golang:1.17.2-buster AS builder
RUN apt-get update -qq \
 && apt-get install -qqy git && \
 mkdir /build

WORKDIR /build

# if you wish to build from upstream, un comment this line, and comment lines below
# RUN git clone --depth 1 https://github.com/solo-io/packer-plugin-arm-image /build

# if you wish to build from upstream, comment from here.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# if you wish to build from upstream, comment up to here.

RUN go build -o packer-plugin-arm-image

FROM docker.io/library/ubuntu:focal

RUN apt-get update -qq \
 && DEBIAN_FRONTEND=noninteractive apt-get install -qqy \
  qemu-user-static \
  unzip \
  wget \
  curl \
  sudo \
 && rm -rf /var/lib/apt/lists/*

ENV PACKER_VERSION 1.7.6

RUN wget https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_amd64.zip -O /tmp/packer.zip && \
  unzip /tmp/packer.zip -d /bin && \
  rm /tmp/packer.zip
WORKDIR /build
COPY entrypoint.sh /entrypoint.sh

COPY --from=builder /build/packer-plugin-arm-image /bin/packer-plugin-arm-image
ENTRYPOINT ["/entrypoint.sh"]
