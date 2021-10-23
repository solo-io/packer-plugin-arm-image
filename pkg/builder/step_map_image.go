package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

var (
	loopRe = regexp.MustCompile("/dev/loop[0-9]+")
)

type stepMapImage struct {
	ImageKey  string
	ResultKey string
}

func (s *stepMapImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is the type we want
	image := state.Get(s.ImageKey).(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("mapping %s", image))

	// Create loopback device
	//   -P (--partscan) creates a partitioned loop device
	//   -f (--find) finds first unused loop device
	//   --show outputs used loop device path
	// Output example:
	//   /dev/loop10
	out, err := exec.Command("losetup", "--show", "-f", "-P", image).CombinedOutput()
	ui.Say(fmt.Sprintf("losetup --show -f -P %s", image))
	if err != nil {
		ui.Error(fmt.Sprintf("error losetup --show -f -P %v: %s", err, string(out)))
		s.Cleanup(state)
		return multistep.ActionHalt
	}
	path := strings.TrimSpace(string(out))
	loop := strings.Split(path, "/")[2]

	// Look for all partitions of created loopback
	var partitions []string
	files, err := os.ReadDir("/dev/")
	if err != nil {
		ui.Error("Couldn't list devices in /dev/")
		s.Cleanup(state)
		return multistep.ActionHalt
	}
	prefix := loop + "p"
	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) {
			partitions = append(partitions, "/dev/"+file.Name())
		}
	}
	partPrefix := "/dev/" + prefix
	// sort by files by partition number, to make sure they match the partition map.
	sort.Slice(partitions, func(i, j int) bool {
		n_i, _ := strconv.Atoi(partitions[i][len(partPrefix):])
		n_j, _ := strconv.Atoi(partitions[j][len(partPrefix):])
		return n_i < n_j
	})

	state.Put(s.ResultKey, partitions)

	return multistep.ActionContinue
}

func (s *stepMapImage) Cleanup(state multistep.StateBag) {
	switch partitions := state.Get(s.ResultKey).(type) {
	case nil:
		return
	case []string:
		if len(partitions) > 0 {
			// Convert /dev/loop10p1 into /dev/loop10
			loop := loopRe.Find([]byte(partitions[0]))
			if loop != nil {
				run(context.TODO(), state, fmt.Sprintf("losetup -d %s", string(loop)))
			}
		}
	}
}
