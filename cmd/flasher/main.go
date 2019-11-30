package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/packer/packer"
	"github.com/solo-io/packer-builder-arm-image/pkg/flasher"
)

func main() {

	device := flag.String("device", "", "device to flash to. leave empty for auto detect")
	image := flag.String("image", "", "image to flash. leave empty for auto detect")
	interactive := flag.Bool("interactive", true, "use interactive mode")
	verify := flag.Bool("verify", true, "verify that image was written")
	flag.Parse()

	flashercfg := flasher.FlashConfig{
		Image:          *image,
		Device:         *device,
		NotInteractive: !*interactive,
		Verify:         *verify,
	}
	// Disable log output by UI
	log.SetOutput(ioutil.Discard)
	var ui packer.Ui = &packer.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
	}

	if os.Geteuid() != 0 {
		ui.Error("Warning: not running as root, this may fail.")
	}

	flshr := flasher.NewFlasher(ui, flashercfg)
	err := flshr.Flash(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(-1)
	} else {
		ui.Say("flashed successfully")
	}
}
