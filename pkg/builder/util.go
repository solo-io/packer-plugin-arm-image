package builder

import (
	"bytes"
	"context"
	"fmt"

	packer_common "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func run(ctx context.Context, state multistep.StateBag, cmds string) error {
	wrappedCommand := state.Get("wrappedCommand").(packer_common.CommandWrapper)
	ui := state.Get("ui").(packer.Ui)

	shellcmd, err := wrappedCommand(cmds)
	if err != nil {
		err := fmt.Errorf("Error creating command '%s': %s", cmds, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return err
	}

	stderr := new(bytes.Buffer)

	cmd := packer_common.ShellCommand(shellcmd)
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
