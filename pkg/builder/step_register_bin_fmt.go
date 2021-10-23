package builder

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const name = "packer-plugin-arm-image"

type stepRegisterBinFmt struct {
	QemuPathKey string
}

// this info can be obtrained with
// /usr/sbin/update-binfmts --display qemu-aarch64
const (
	mask               = `\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff`
	qemu_arm_magic     = `\x7f\x45\x4c\x46\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00`
	qemu_aarch64_magic = `\x7f\x45\x4c\x46\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00`
)

func (s *stepRegisterBinFmt) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	ui := state.Get("ui").(packer.Ui)
	qemu := state.Get(s.QemuPathKey).(string)

	ui.Say("Registering " + qemu + " with binfmt_misc")

	registerstring_prefix := []byte{':'}
	registerstring_prefix = append(registerstring_prefix, []byte(name)...)
	registerstring_prefix = append(registerstring_prefix, ':', 'M', ':', ':')
	if strings.Contains(qemu, "64") {
		registerstring_prefix = append(registerstring_prefix, qemu_aarch64_magic...)
	} else {
		registerstring_prefix = append(registerstring_prefix, qemu_arm_magic...)
	}

	registerstring_prefix = append(registerstring_prefix, ':')
	registerstring_prefix = append(registerstring_prefix, []byte(mask)...)
	registerstring_prefix = append(registerstring_prefix, ':')
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
