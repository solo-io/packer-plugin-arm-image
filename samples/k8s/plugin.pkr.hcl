packer {
  required_plugins {
    arm-image = {
      version = ">= 0.2.5"
      source  = "github.com/solo-io/arm-image"
    }
  }
}
