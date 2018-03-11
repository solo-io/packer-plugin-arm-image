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
	extrata := config.LastPartitionExtraSize

	ui.Say(fmt.Sprintf("Resizing the last partition %v.", extrata))

	if extrata == 0 {
		return multistep.ActionContinue
	}

	// resizer image
	err := s.enlargeImage(state, imagefile, int64(extrata))
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
	extrasector := extrata >> SectorShift
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
	part.SetLBALen(part.GetLBALen() + uint32(extrasector))

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

func (s *stepResizeLastPart) enlargeImage(state multistep.StateBag, imagefile string, extrata int64) error {
	stat, err := os.Stat(imagefile)
	if err != nil {
		return fmt.Errorf("can't stat file  %v", err)
	}
	return os.Truncate(imagefile, stat.Size()+extrata)
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
