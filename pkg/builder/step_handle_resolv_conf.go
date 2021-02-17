package builder

// taken from here: https://github.com/hashicorp/packer/blob/81522dced0b25084a824e79efda02483b12dc7cd/builder/amazon/chroot/step_chroot_provision.go

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// stepHandleResolvConf provisions the instance within a chroot.
type stepHandleResolvConf struct {
	ChrootKey string
	Delete    bool
}

func (s *stepHandleResolvConf) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	mountPath := state.Get(s.ChrootKey).(string)
	ui := state.Get("ui").(packer.Ui)

	const origResolvConf = "/etc/resolv.conf"
	destResolvConf := filepath.Join(mountPath, origResolvConf)

	if s.Delete {
		err := os.Remove(destResolvConf)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		// copy file over:
		err := copyFile(destResolvConf, origResolvConf)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepHandleResolvConf) Cleanup(state multistep.StateBag) {}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
