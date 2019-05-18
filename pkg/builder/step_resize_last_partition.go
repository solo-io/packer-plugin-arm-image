package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
	targetSize := config.TargetImageSize
	extrata := config.LastPartitionExtraSize // legacy way to specify extension

	if extrata == 0 && targetSize == 0 {
		return multistep.ActionContinue
	}

	stat, err := os.Stat(imagefile)
	if err != nil {
		ui.Error(fmt.Sprintf("Cannot stat() image file: %v", err))
		return multistep.ActionHalt
	}

	if targetSize > 0 && targetSize < uint64(stat.Size()) {
		ui.Error(fmt.Sprintf("target_image_size (%v) is smaller than actual image size (%v). Cannot shrink image.",
			targetSize, stat.Size()))
		return multistep.ActionHalt
	}
	if targetSize > 0 && extrata > 0 {
		ui.Say("both last_partition_extra_size and target_image_size was specified - ignoring last_partition_extra_size")
		extrata = uint64(stat.Size()) - targetSize
	}

	ui.Say("Resizing the last partition.")
	err = os.Truncate(imagefile, int64(extrata)+stat.Size())
	// resizer image
	if err != nil {
		ui.Error(fmt.Sprintf("Error enlarging image file %v", err))
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
	extrasector := uint32(extrata >> SectorShift)
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
