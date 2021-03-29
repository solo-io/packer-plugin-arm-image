package builder

import (
	"context"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepRegisterBinFmt struct {
	QemuPathKey string
}

func (s *stepRegisterBinFmt) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	ui := state.Get("ui").(packer.Ui)
	qemu := state.Get(s.QemuPathKey).(string)

	// registerstring := `:packer-builder-arm-image:M::\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:/usr/bin/qemu-arm-static:`

	registerstring_prefix := []byte{':', 'p', 'a', 'c', 'k', 'e', 'r', '-', 'b', 'u', 'i', 'l', 'd', 'e', 'r', '-', 'a', 'r', 'm', '-', 'i', 'm', 'a', 'g', 'e', ':', 'M', ':', ':', '\\', 'x', '7', 'f', 'E', 'L', 'F', '\\', 'x', '0', '1', '\\', 'x', '0', '1', '\\', 'x', '0', '1', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '2', '\\', 'x', '0', '0', '(', '\\', 'x', '0', '0', ':', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', '0', '0', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'e', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', ':'}
	registerstring := append(registerstring_prefix, ([]byte(qemu))...)
	registerstring = append(registerstring, ':')
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
