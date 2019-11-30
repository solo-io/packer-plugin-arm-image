package builder

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepMapImage struct {
	ImageKey  string
	ResultKey string
}

func (s *stepMapImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is they type we want
	image := state.Get(s.ImageKey).(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("mappping %s", image))
	// if run(state, fmt.Sprintf(
	//	"kpartx -s -a %s",
	//	image)) != nil {
	//	return multistep.ActionHalt
	//}

	out, err := exec.Command("kpartx", "-s", "-a", "-v", image).CombinedOutput()
	ui.Say(fmt.Sprintf("kpartx -s -a -v %s", image))

	// out, err := exec.Command("kpartx", "-l", image).CombinedOutput()
	// ui.Say(fmt.Sprintf("kpartx -l: %s", string(out)))
	if err != nil {
		ui.Error(fmt.Sprintf("error kaprts -l %v: %s", err, string(out)))
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	// get the loopback device for the partitions
	// kpartx -l output looks like this:
	/*
		loop2p1 : 0 85045 /dev/loop2 8192
		loop2p2 : 0 3534848 /dev/loop2 94208
	*/
	/*
		  kpartx -a -v output looks like this:

			add map loop20p1 (254:22): 0 88262 linear 7:20 8192
			add map loop20p2 (254:23): 0 3538944 linear 7:20 98304
	*/
	lines := strings.Split(string(out), "\n")

	var partitions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		device := strings.Split(string(line), " ")
		if len(device) != 9 {
			ui.Error("bad kpartx output: " + string(out))
			s.Cleanup(state)
			return multistep.ActionHalt
		}
		partitions = append(partitions, "/dev/mapper/"+device[2])
	}

	state.Put(s.ResultKey, partitions)

	return multistep.ActionContinue
}

func (s *stepMapImage) Cleanup(state multistep.StateBag) {
	image := state.Get(s.ImageKey).(string)
	run(context.TODO(), state, fmt.Sprintf("kpartx -d %s", image))
}
