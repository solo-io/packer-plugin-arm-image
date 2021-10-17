package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mattn/go-tty"
	"github.com/solo-io/packer-plugin-arm-image/pkg/flasher"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

// realMain is executed from main and returns the exit status to exit with.
func realMain() int {
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
	var ui *packer.BasicUi = &packer.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
	}

	if !flashercfg.NotInteractive {
		if TTY, err := tty.Open(); err != nil {
			fmt.Fprintf(os.Stderr, "No tty available: %s\n", err)
		} else {
			ui.TTY = TTY
			defer TTY.Close()
		}
	}

	if os.Geteuid() != 0 {
		ui.Error("Warning: not running as root, this may fail.")
	}

	flshr := flasher.NewFlasher(ui, flashercfg)
	err := flshr.Flash(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return -1
	} else {
		ui.Say("flashed successfully")
	}
	return 0
}
