package builder

// This file was copied and modified from aws chroot builder.
import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepMountCleanup mounts the attached device.
//
// Produces:
//   mount_extra_cleanup CleanupFunc - To perform early cleanup
type StepMountCleanup struct {
}

func (s *StepMountCleanup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *StepMountCleanup) Cleanup(state multistep.StateBag) {
	mountPath := state.Get("mount_path").(string)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("fuser -k " + mountPath)
	run(context.TODO(), state, "fuser -k "+mountPath+" || exit 0")
}
