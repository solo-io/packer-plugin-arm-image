package builder

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepResizeFs struct {
	PartitionsKey string
}

func (s *stepResizeFs) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	// Read our value and assert that it is they type we want
	partitions := state.Get(s.PartitionsKey).([]string)
	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("partitions: %v", partitions))

	p := partitions[len(partitions)-1]
	err := s.e2fsck(ctx, wrappedCommand, p)
	if err != nil {
		err := fmt.Errorf("Error e2fsck command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = s.resize(ctx, wrappedCommand, p)

	if err != nil {
		err := fmt.Errorf("Error creating resize command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepResizeFs) e2fsck(ctx context.Context, wrappedCommand CommandWrapper, dev string) error {
	e2fsckCommand, err := wrappedCommand(fmt.Sprintf("e2fsck -y -f %s", dev))
	if err != nil {
		return err
	}

	cmd := ShellCommand(ctx, e2fsckCommand)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"Error e2fsck: %s\nStderr: %s", err, stderr.String())
		return err
	}
	return nil
}

func (s *stepResizeFs) resize(ctx context.Context, wrappedCommand CommandWrapper, dev string) error {

	reizeCommand, err := wrappedCommand(fmt.Sprintf("resize2fs %s", dev))
	if err != nil {
		return err
	}

	cmd := ShellCommand(ctx, reizeCommand)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"Error resizing: %s\nStderr: %s", err, stderr.String())
		return err
	}
	return nil
}

func (s *stepResizeFs) Cleanup(state multistep.StateBag) {
}
