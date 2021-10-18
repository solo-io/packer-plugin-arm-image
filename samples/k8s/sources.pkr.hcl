source "arm-image" "k8s" {
  iso_url = "https://downloads.raspberrypi.org/raspios_lite_arm64/images/raspios_lite_arm64-2021-05-28/2021-05-07-raspios-buster-arm64-lite.zip"
  iso_checksum = "sha256:868cca691a75e4280c878eb6944d95e9789fa5f4bbce2c84060d4c39d057a042"
  output_filename = "../output-arm-image/k8s.img"
  target_image_size = 3*1024*1024*1024
  qemu_binary = "qemu-aarch64-static"
}
