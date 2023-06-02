//go:build darwin
// +build darwin

package builder

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepMapImage struct {
	ImageKey  string
	ResultKey string
}

var (
	whitespaceRe = regexp.MustCompile(`\s+`)
	diskRe       = regexp.MustCompile(`(?m)^(/dev/disk[0-9]+)[a-z][0-9]+$`)
)

func (s *stepMapImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// Read our value and assert that it is the type we want
	image := state.Get(s.ImageKey).(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("mapping %s", image))

	// Attach disk image
	//   -nomount	same as -mount suppressed
	//   -imagekey	diskimage-class=CRawDiskImage
	// Output example:
	//   /dev/disk2 	FDisk_partition_scheme
	//   /dev/disk2s1	Windows_FAT_32
	//   /dev/disk2s2	Linux
	out, err := exec.Command("hdiutil", "attach",
		"-imagekey", "diskimage-class=CRawDiskImage",
		"-nomount", image).CombinedOutput()
	cmd := fmt.Sprintf("hdiutil attach -imagekey diskimage-class=CRawDiskImage -nomount %s", image)
	ui.Say(cmd)
	if err != nil {
		ui.Error(fmt.Sprintf("error %s %v: %s", cmd, err, string(out)))
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	// Look for all partitions of created loopback
	var partitions []string
	lines := strings.Split(string(out), "\n")

	// make sure /dev/disk2 is always the first line,
	// also to make sure disks match the partition map.
	sort.Strings(lines)
	for _, l := range lines {
		split := whitespaceRe.Split(strings.TrimSpace(l), -1)
		if len(split) != 2 {
			continue
		}
		partition := split[0]
		if !diskRe.MatchString(partition) {
			continue
		}
		partitions = append(partitions, partition)
	}

	state.Put(s.ResultKey, partitions)

	return multistep.ActionContinue
}

func (s *stepMapImage) Cleanup(state multistep.StateBag) {
	switch partitions := state.Get(s.ResultKey).(type) {
	case nil:
		return
	case []string:
		if len(partitions) > 0 {
			// Convert /dev/disk2s1 into /dev/disk2
			matches := diskRe.FindAllStringSubmatch(partitions[0], -1)
			if len(matches) == 0 {
				// must not happen
				return
			}
			if len(matches[0]) == 0 {
				// must not happen
				return
			}
			disk := matches[0][1]
			run(context.TODO(), state, fmt.Sprintf(
				"hdiutil eject -force %s", disk))
		}
	}
}
