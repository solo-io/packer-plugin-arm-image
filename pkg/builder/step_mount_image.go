package builder

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepMountImage struct {
	PartitionsKey    string
	ResultKey        string
	MountPath        string
	GeneratedDataKey string
	mountpoints      []string
}

func (s *stepMountImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
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

	if len(s.MountPath) > 0 {
		err := os.MkdirAll(s.MountPath, os.ModePerm)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		tempDir, err := ioutil.TempDir("", "armimg-")
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		s.MountPath = tempDir
	}
	log.Println("mounting to", s.MountPath)

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

		mntpnt := filepath.Join(s.MountPath, mntAndPart.mnt)

		ui.Message(fmt.Sprintf("Mounting: %s", mntAndPart.part))

		err := run(ctx, state, fmt.Sprintf(
			"mount %s %s",
			mntAndPart.part, mntpnt))
		if err != nil {
			return multistep.ActionHalt
		}

		s.mountpoints = append(s.mountpoints, mntpnt)
	}

	state.Put(s.ResultKey, s.MountPath)

	updateGeneratedData(state, s.GeneratedDataKey, s.MountPath)

	return multistep.ActionContinue
}

func (s *stepMountImage) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if s.MountPath != "" {
		for _, mntpnt := range reverse(s.mountpoints) {
			run(context.TODO(), state, "umount "+mntpnt)
		}
		s.mountpoints = nil
		// DO NOT do remove all here! if dev fails to umount it would be undesirable.
		err := os.Remove(s.MountPath)
		if err != nil {
			ui.Error(err.Error())
		}

		s.MountPath = ""
	}
}

func reverse(numbers []string) []string {
	newNumbers := make([]string, len(numbers))
	for i, j := 0, len(numbers)-1; i <= j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
}

func updateGeneratedData(state multistep.StateBag, key string, value string) {
	generatedData, found := state.GetOk("generated_data")

	if found {
		generatedData.(map[string]interface{})[key] = value
	} else {
		state.Put("generated_data", map[string]interface{}{key: value})
	}
}
