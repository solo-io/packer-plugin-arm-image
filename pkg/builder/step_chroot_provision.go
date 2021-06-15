package builder

// taken from here: https://github.com/hashicorp/packer/blob/81522dced0b25084a824e79efda02483b12dc7cd/builder/amazon/chroot/step_chroot_provision.go

import (
	"context"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/chroot"
	packer_common_common "github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepChrootProvision provisions the instance within a chroot.
type StepChrootProvision struct {
	ChrootKey string
}

func (s *StepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	hook := state.Get("hook").(packer.Hook)
	mountPath := state.Get(s.ChrootKey).(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(packer_common_common.CommandWrapper)

	// Create our communicator
	comm := &chroot.Communicator{
		Chroot:     mountPath,
		CmdWrapper: wrappedCommand,
	}

	// Provision
	log.Println("Running the provision hook")
	if err := hook.Run(ctx, packer.HookProvision, ui, comm, nil); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepChrootProvision) Cleanup(state multistep.StateBag) {}
