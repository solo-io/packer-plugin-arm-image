This sample creates a WiFi extender.
plug your raspberry pi to your router, and it will create a wifi that extends the router.
build like so:


```
cd to/samples
docker run \
  --rm \
  --privileged \
  -v ${PWD}:/build:ro \
  -v ${PWD}/packer_cache:/build/packer_cache \
  -v ${PWD}/output-arm-image:/build/output-arm-image \
  -e PACKER_CACHE_DIR=/build/packer_cache \
  -w /build/hostapd \
  quay.io/solo-io/packer-builder-arm-image:v0.1.4.5 build .
```