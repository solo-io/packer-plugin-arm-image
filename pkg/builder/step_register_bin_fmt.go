package builder

import (
	"os"

	"github.com/hashicorp/packer/packer"

	"github.com/mitchellh/multistep"
)

type stepRegisterBinFmt struct {
	destQemu string
}

func (s *stepRegisterBinFmt) Run(state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	ui := state.Get("ui").(packer.Ui)

	// registerstring := `:packer-builder-arm-image:M::\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:/usr/bin/qemu-arm-static:`
	registerstring := []byte{':', 'p', 'a', 'c', 'k', 'e', 'r', '-', 'b', 'u', 'i', 'l', 'd', 'e', 'r', '-', 'a', 'r', 'm', '-', 'i', 'm', 'a', 'g', 'e', ':', 'M', ':', ':', '\\', 'x', '7', 'f', 'E', 'L', 'F', '\\', 'x', '0', '1', '\\', 'x', '0', '1', '\\', 'x', '0', '1', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '2', '\\', 'x', '0', '0', '(', '\\', 'x', '0', '0', ':', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', '0', '0', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'e', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', ':', '/', 'u', 's', 'r', '/', 'b', 'i', 'n', '/', 'q', 'e', 'm', 'u', '-', 'a', 'r', 'm', '-', 's', 't', 'a', 't', 'i', 'c', ':'}
	f, err := os.OpenFile("/proc/sys/fs/binfmt_misc/register", os.O_RDWR, 0)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer f.Close()
	_, err = f.Write(registerstring)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepRegisterBinFmt) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	f, err := os.OpenFile("/proc/sys/fs/binfmt_misc/packer-builder-arm-image", os.O_RDWR, 0)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	defer f.Close()
	_, err = f.WriteString("-1")
}
