source "arm-image" "hostapd" {
  iso_url = "https://downloads.raspberrypi.org/raspbian_lite/images/raspbian_lite-2020-02-14/2020-02-13-raspbian-buster-lite.zip"
  iso_checksum = "sha256:12ae6e17bf95b6ba83beca61e7394e7411b45eba7e6a520f434b0748ea7370e8"
  output_filename = "../output-arm-image/hostapd.img"
  target_image_size = 3*1024*1024*1024
}
