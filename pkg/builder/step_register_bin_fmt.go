package builder

import (
	"context"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const name = "packer-plugin-arm-image"

type stepRegisterBinFmt struct {
	QemuPathKey string
}

func (s *stepRegisterBinFmt) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	ui := state.Get("ui").(packer.Ui)
	qemu := state.Get(s.QemuPathKey).(string)

	ui.Say("Registering " + qemu + " with binfmt_misc")

	registerstring_prefix := []byte{':'}
	registerstring_prefix = append(registerstring_prefix, ([]byte(name))...)
	registerstring_prefix = append(registerstring_prefix, []byte{':', 'M', ':', ':', '\\', 'x', '7', 'f', 'E', 'L', 'F', '\\', 'x', '0', '1', '\\', 'x', '0', '1', '\\', 'x', '0', '1', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '0', '\\', 'x', '0', '2', '\\', 'x', '0', '0', '(', '\\', 'x', '0', '0', ':', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', '0', '0', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'e', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', '\\', 'x', 'f', 'f', ':'}...)
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

	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/" + name); os.IsNotExist(err) {
		ui.Error("binfmt_misc registration failed")
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepRegisterBinFmt) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	f, err := os.OpenFile("/proc/sys/fs/binfmt_misc/"+name, os.O_RDWR, 0)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	defer f.Close()
	_, err = f.WriteString("-1\n")
	if err != nil {
		ui.Error("Failed de-registering binfmt_misc" + err.Error())
	}
}
