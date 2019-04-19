package builder

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepMountImage struct {
	PartitionsKey string
	ResultKey     string
	tempdir       string
	mountpoints   []string
}

func (s *stepMountImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)

	// Read our value and assert that it is they type we want
	partitions := state.Get(s.PartitionsKey).([]string)
	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("partitions: %v", partitions))

	// assume first one is boot and second one is root!
	if len(partitions) != len(config.ImageMounts) {
		ui.Error(fmt.Sprintf("error different of partitions than expected %v", len(partitions)))
		return multistep.ActionHalt
	}

	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.tempdir = tempdir

	mountsAndPartitions := make([]struct{ part, mnt string }, len(partitions))
	for i := range partitions {
		mountsAndPartitions[i].part = partitions[i]
		mountsAndPartitions[i].mnt = config.ImageMounts[i]
	}

	// sort so we mount with the right order
	// sort that / is mounted before /boot
	sort.Slice(mountsAndPartitions, func(i, j int) bool { return mountsAndPartitions[i].mnt < mountsAndPartitions[j].mnt })

	for _, mntAndPart := range mountsAndPartitions {
		if mntAndPart.mnt == "" {
			continue
		}

		mntpnt := filepath.Join(s.tempdir, mntAndPart.mnt)

		ui.Message(fmt.Sprintf("Mounting: %s", mntAndPart.part))

		err := run(state, fmt.Sprintf(
			"mount %s %s",
			mntAndPart.part, mntpnt))
		if err != nil {
			return multistep.ActionHalt
		}

		s.mountpoints = append(s.mountpoints, mntpnt)
	}

	state.Put(s.ResultKey, tempdir)
	return multistep.ActionContinue
}

func (s *stepMountImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if s.tempdir != "" {
		for _, mntpnt := range reverse(s.mountpoints) {
			run(state, "umount "+mntpnt)
		}
		s.mountpoints = nil
		// DO NOT do remove all here! if dev fails to umount it would be undesirable.
		err := os.Remove(s.tempdir)
		if err != nil {
			ui.Error(err.Error())
		}

		s.tempdir = ""
	}
}

func reverse(numbers []string) []string {
	newNumbers := make([]string, len(numbers))
	for i, j := 0, len(numbers)-1; i <= j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
}
