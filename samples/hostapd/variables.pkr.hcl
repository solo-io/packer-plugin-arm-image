variable "wifi_ssid" {
  type = string
}
variable "wifi_psk" {
  type = string
  sensitive = true
}
variable "image_home_dir" {
  type = string
  default = "/home/pi"
}
variable "local_ssh_public_key" {
  type = string
  default = "~/.ssh/id_rsa.pub"
}

locals {
  ssh_key = "${pathexpand(var.local_ssh_public_key)}"
}
