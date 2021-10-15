package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/solo-io/packer-plugin-arm-image/pkg/image"
	"github.com/solo-io/packer-plugin-arm-image/pkg/utils"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCopyImage struct {
	FromKey, ResultKey string
	ImageOpener        image.ImageOpener
	ui                 packer.Ui
}

func (s *stepCopyImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	fromFile := state.Get(s.FromKey).(string)
	config := state.Get("config").(*Config)
	s.ui = state.Get("ui").(packer.Ui)
	s.ui.Say("Copying source image.")

	outputDir := filepath.Dir(config.OutputFile)
	imageName := filepath.Base(config.OutputFile)

	err := s.copy(ctx, state, fromFile, outputDir, imageName)
	if err != nil {
		s.ui.Error(fmt.Sprintf("%v", err))
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, config.OutputFile)
	return multistep.ActionContinue
}

func (s *stepCopyImage) Cleanup(state multistep.StateBag) {
}

func (s *stepCopyImage) copy_progress(ctx context.Context, state multistep.StateBag, dst io.Writer, src image.Image) error {
	ui := state.Get("ui").(packer.Ui)

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	defer close(done)
	go func() {
		defer cancel()
		for {
			select {
			case <-time.After(time.Second):
				if _, ok := state.GetOk(multistep.StateCancelled); ok {
					ui.Say("Interrupt received. Cancelling copy...")
					break
				}
			case <-done:
				return
			}
		}
	}()

	_, err := utils.CopyWithProgress(ctx, ui, dst, src)
	return err

}

func (s *stepCopyImage) copy(ctx context.Context, state multistep.StateBag, src, dir, filename string) error {

	srcf, err := s.ImageOpener.Open(src)
	if err != nil {
		return err
	}
	defer srcf.Close()

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	dstf, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer dstf.Close()

	err = s.copy_progress(ctx, state, dstf, srcf)

	if err != nil {
		return err
	}

	return nil
}
