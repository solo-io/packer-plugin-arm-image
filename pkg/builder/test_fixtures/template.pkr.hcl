source "arm-image" "test" {
  iso_url = "test_fixtures/img.bin.gz"
  iso_checksum = "none"
  output_filename = "img.delete"
  image_mounts = ["/"]
}

build {
  sources = [
    "source.arm-image.test"
  ]
  # install hostapd, bridge utils, and other utilities.
  provisioner "shell" {
    inline = [
      "echo hello world"
    ]
  }
}
