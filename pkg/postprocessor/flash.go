package postprocessor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/solo-io/packer-builder-arm-image/pkg/utils"
)

const BlockSize = 512

type FlashConfig struct {
	Device      string `mapstructure:"device"`
	Interactive bool   `mapstructure:"interactive"`
	Verify      bool   `mapstructure:"verify"`
}

type Flasher struct {
	config FlashConfig
}

func NewFlasher() packer.PostProcessor {
	return &Flasher{}
}

func (f *Flasher) Configure(cfgs ...interface{}) error {
	err := config.Decode(&f.config, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, cfgs...)
	if err != nil {
		return err
	}
	return nil
}

func (f *Flasher) PostProcess(ui packer.Ui, ain packer.Artifact) (a packer.Artifact, keep bool, err error) {
	inputfiles := ain.Files()
	if len(inputfiles) != 1 {
		return nil, false, errors.New("ambigous images")
	}
	imageToFlash := inputfiles[0]

	dev, err := f.getDevice(ui)
	if err != nil {
		return nil, false, err
	}

	if f.config.Interactive {
		answer, err := ui.Ask(fmt.Sprintf("Going to flash to %s. Are you sure?", dev.Device))
		if err != nil {
			return nil, false, err
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if !strings.HasPrefix("yes", answer) {
			return nil, false, errors.New("canceled by user")
		}
	}

	err = f.unmount(ui, dev)
	if err != nil {
		return nil, false, err
	}
	err = f.flash(ui, imageToFlash, dev)
	if err != nil {
		return nil, false, err
	}
	return nil, false, nil

}

func (f *Flasher) unmount(ui packer.Ui, device *utils.Device) error {
	for _, mntpnt := range device.Mountpoints {
		ui.Say("unmounting " + mntpnt)
		err := exec.Command("umount", mntpnt).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

type WriterSeeker interface {
	io.Seeker
	io.Writer
}

func (f *Flasher) flash(ui packer.Ui, file string, device *utils.Device) error {
	finfo, err := os.Stat(file)
	if err != nil {
		return err
	}
	fsize := finfo.Size()

	input, err := os.Open(file)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(device.Device, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer output.Close()
	progress := utils.NewProgressWriterWithSize(uint64(fsize))
	rdr := io.TeeReader(input, progress)

	copyCompleteCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(output, rdr)
		copyCompleteCh <- err
	}()

	progressTicker := time.NewTicker(15 * time.Second)
	defer progressTicker.Stop()

	for {
		select {
		case err := <-copyCompleteCh:

			return err
		case <-progressTicker.C:
			progress := progress.Progress()
			if progress.MBytesPerSecond >= 0 {
				ui.Message(fmt.Sprintf("Copy speed: %7.2f MB/s", progress.MBytesPerSecond))
			}
			if progress.PercentDone > 0 {
				ui.Message(fmt.Sprintf("Copy progress: %3.2f%%", progress.PercentDone))
			}
			/* TODO use steps
			case <-ctx.Done():
				l.Stop()
				return errors.New("interrupted")
			case <-time.After(1 * time.Second):
				if _, ok := state.GetOk(multistep.StateCancelled); ok {
					ui.Say("Interrupt received. Cancelling copy...")
					l.Stop()
					return errors.New("interrupted")
				}
			*/
		}
	}
	return nil
}

func (f *Flasher) getDevice(ui packer.Ui) (*utils.Device, error) {

	detachables, err := utils.GetDetachableDevices()
	if err != nil {
		return nil, err
	}
	if len(detachables) == 0 {
		return nil, errors.New("no devices")
	}

	if len(f.config.Device) != 0 {
		for _, d := range detachables {
			if d.Device == f.config.Device {
				return &d, nil
			}
		}
	}

	if !f.config.Interactive {
		if len(detachables) != 1 {
			return nil, errors.New("ambigous device")

		}
		return &detachables[0], nil
	} else {
		question := "Which device should we choose?:\n"
		for i, d := range detachables {
			question += fmt.Sprint("%d. %s (%s)", i+1, d.Device, d.Name)
		}
		answer, err := ui.Ask(question)
		if err != nil {
			return nil, err
		}
		i, err := strconv.Atoi(answer)
		if err != nil {
			return nil, err
		}
		i = i - 1
		if i < 0 || i > len(detachables) {
			return nil, errors.New("invalid selection")
		}
		return &detachables[i], nil
	}

	return nil, errors.New("no device")
}
