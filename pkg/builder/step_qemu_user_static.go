package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
)

const qemyBinary = "/usr/bin/qemu-arm-static"

type stepQemuUserStatic struct {
	ChrootKey string
	destQemu  string
}

func (s *stepQemuUserStatic) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	chrootDir := state.Get(s.ChrootKey).(string)

	srcqemu := qemyBinary
	s.destQemu = filepath.Join(chrootDir, srcqemu)
	err := run(state, fmt.Sprintf("cp %s %s", srcqemu, s.destQemu))
	if err != nil {
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *stepQemuUserStatic) Cleanup(state multistep.StateBag) {
	if s.destQemu != "" {
		os.Remove(s.destQemu)
	}
}
