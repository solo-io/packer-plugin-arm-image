package builder

import (
	"bytes"
	"context"
	"fmt"

	packer_common_common "github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func run(ctx context.Context, state multistep.StateBag, cmds string) error {
	wrappedCommand := state.Get("wrappedCommand").(packer_common_common.CommandWrapper)
	ui := state.Get("ui").(packer.Ui)

	shellcmd, err := wrappedCommand(cmds)
	if err != nil {
		err := fmt.Errorf("Error creating command '%s': %s", cmds, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return err
	}

	stderr := new(bytes.Buffer)

	cmd := packer_common_common.ShellCommand(shellcmd)
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
