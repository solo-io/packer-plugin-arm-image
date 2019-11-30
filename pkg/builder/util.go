package builder

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer"

	"github.com/hashicorp/packer/helper/multistep"
)

func run(ctx context.Context, state multistep.StateBag, cmds string) error {
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)
	ui := state.Get("ui").(packer.Ui)

	shellcmd, err := wrappedCommand(cmds)
	if err != nil {
		err := fmt.Errorf("Error creating command '%s': %s", cmds, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return err
	}

	stderr := new(bytes.Buffer)

	cmd := ShellCommand(ctx, shellcmd)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"Error executing command '%s': %s\nStderr: %s", cmds, err, stderr.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return err
	}
	return nil
}
