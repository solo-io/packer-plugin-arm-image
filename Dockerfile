FROM golang:buster AS builder
RUN apt-get update -qq \
 && apt-get install -qqy git 
RUN git clone --depth 1 https://github.com/solo-io/packer-builder-arm-image /build
WORKDIR /build
RUN go build

FROM ubuntu:eoan
ENV PACKER_VERSION 1.4.5
ADD https://releases.hashicorp.com/packer/${PACKER_VERSION}/packer_${PACKER_VERSION}_linux_amd64.zip /tmp/packer.zip
COPY --from=builder /build/packer-builder-arm-image /bin/packer-builder-arm-image
RUN apt-get update -qq \
 && DEBIAN_FRONTEND=noninteractive apt-get install -qqy \
  qemu-user-static \
  kpartx \
  unzip \
  wget \
  curl \
  sudo \
 && rm -rf /var/lib/apt/lists/* \
 && unzip /tmp/packer.zip -d /bin && rm /tmp/packer.zip
WORKDIR /build
ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
