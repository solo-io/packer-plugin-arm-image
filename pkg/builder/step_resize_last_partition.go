package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/rekby/mbr"
)

// sector size is 512 bytes
const SectorShift = 9

type stepResizeLastPart struct {
	FromKey string
}

func (s *stepResizeLastPart) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	imagefile := state.Get(s.FromKey).(string)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	extraSize := int64(config.LastPartitionExtraSize) // legacy way to resize last partition
	targetSize := int64(config.TargetImageSize)

	if extraSize <= 0 && targetSize <= 0 {
		return multistep.ActionContinue
	}

	stat, err := os.Stat(imagefile)
	if err != nil {
		ui.Error(fmt.Sprintf("Cannot stat() image file: %v", err))
		return multistep.ActionHalt
	}

	currentSize := stat.Size()
	if targetSize > 0 {
		if targetSize < currentSize {
			ui.Error(fmt.Sprintf("Cannot shrink partition, current size is %v, new size is %v",
				currentSize, targetSize))
			return multistep.ActionHalt
		}

		if targetSize == currentSize {
			return multistep.ActionContinue
		}

		ui.Say(fmt.Sprintf("Growing partition to %v M (%v bytes)", targetSize/1024/1024, targetSize))
		extraSize = targetSize - currentSize
	} else {
		ui.Say(fmt.Sprintf("Growing partition with %v M (%v bytes)", extraSize/1024/1024, extraSize))
		targetSize = currentSize + extraSize
	}

	// resize image
	err = os.Truncate(imagefile, targetSize)
	if err != nil {
		ui.Error(fmt.Sprintf("Error growing image file %v", err))
		return multistep.ActionHalt
	}

	// resize the last partition
	mbrp, err := s.getMbr(imagefile)
	if err != nil {
		ui.Error(fmt.Sprintf("Error retreiving mbr %v", err))
		return multistep.ActionHalt
	}
	partitions := mbrp.GetAllPartitions()

	if len(partitions) == 0 {
		ui.Error("no partitions!")
		return multistep.ActionHalt
	}

	var part *mbr.MBRPartition
	for _, potentialpart := range partitions {
		if !potentialpart.IsEmpty() {
			part = potentialpart
		}
	}

	if part == nil {
		ui.Error(fmt.Sprintf("no partition %v", *mbrp))
		return multistep.ActionHalt
	}
	extrasector := uint32(extraSize >> SectorShift)
	part.SetLBALen(part.GetLBALen() + extrasector)

	f, err := os.OpenFile(imagefile, os.O_RDWR|os.O_SYNC, 0600)
	if err != nil {
		ui.Error(fmt.Sprintf("Can't open image for writing %v", err))
		return multistep.ActionHalt
	}

	defer f.Close()
	err = mbrp.Write(f)
	if err != nil {
		ui.Error(fmt.Sprintf("Can't write mbr  %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepResizeLastPart) getMbr(imagefile string) (*mbr.MBR, error) {

	disk, err := os.Open(imagefile)
	if err != nil {
		return nil, err
	}
	defer disk.Close()

	return mbr.Read(disk)

}

func (s *stepResizeLastPart) Cleanup(state multistep.StateBag) {
}

type zeroreader struct{}

func (*zeroreader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
