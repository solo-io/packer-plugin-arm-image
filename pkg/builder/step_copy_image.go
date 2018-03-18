package builder

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/solo-io/packer-builder-arm-image/pkg/utils"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	filetype "gopkg.in/h2non/filetype.v1"
	"gopkg.in/h2non/filetype.v1/matchers"

	"github.com/ulikunitz/xz"
)

type stepCopyImage struct {
	FromKey, ResultKey string
	ui                 packer.Ui
}

func (s *stepCopyImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	fromFile := state.Get(s.FromKey).(string)
	config := state.Get("config").(*Config)
	s.ui = state.Get("ui").(packer.Ui)
	s.ui.Say("Copying source image.")

	imageName := "image"

	dstfile := filepath.Join(config.OutputDir, imageName)
	err := s.copy(ctx, state, fromFile, config.OutputDir, imageName)
	if err != nil {
		s.ui.Error(fmt.Sprintf("%v", err))
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, dstfile)
	return multistep.ActionContinue
}

func (s *stepCopyImage) Cleanup(state multistep.StateBag) {
}

func (s *stepCopyImage) open(fpath string) (io.ReadCloser, error) {
	t, _ := filetype.MatchFile(fpath)

	f, err := os.Open(fpath)

	switch t {
	case matchers.TypeZip:
		s.ui.Say("Image is a zip file.")
		return s.openzip(f)
	case matchers.TypeXz:
		s.ui.Say("Image is a xz file.")
		return s.openxz(f)
	default:
		return f, err
	}

}

func (s *stepCopyImage) openzip(f *os.File) (io.ReadCloser, error) {
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	fstat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	r, err := zip.NewReader(f, fstat.Size())
	if err != nil {
		return nil, err
	}

	if len(r.File) != 1 {
		return nil, errors.New("support for only zip files with one file.")
	}

	zippedfile := r.File[0]
	s.ui.Say("Unzipping " + zippedfile.Name)
	zippedfileReader, err := zippedfile.Open()
	if err != nil {
		return nil, err
	}

	//transfer ownership
	mc := &multiCloser{zippedfileReader, []io.Closer{zippedfileReader, f}}
	f = nil

	return mc, nil
}

func (s *stepCopyImage) xzFastlane(f *os.File) (io.ReadCloser, error) {

	xzcat := exec.Command("xzcat")

	// fast path, use xzcat
	xzcat.Stdin = f
	r, err := xzcat.StdoutPipe()

	if err != nil {
		return nil, err
	}
	if err := xzcat.Start(); err != nil {
		return nil, err
	}

	go func() {
		xzcat.Wait()
	}()

	//	mc := &multiCloser{r, []io.Closer{f}}
	return r, nil

}

func (s *stepCopyImage) openxz(f *os.File) (io.ReadCloser, error) {
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	// check if available:
	if exec.Command("which", "xzcat").Run() == nil {
		ret, err := s.xzFastlane(f)
		if err == nil {
			f = nil
			return ret, err
		}
	}
	// slow lane here
	r, err := xz.NewReader(f)
	if err != nil {
		return nil, err
	}

	//transfer ownership
	mc := &multiCloser{r, []io.Closer{f}}
	f = nil

	return mc, nil
}

type multiCloser struct {
	io.Reader
	c []io.Closer
}

func (n *multiCloser) Close() error {
	for _, c := range n.c {
		c.Close()
	}
	return nil
}

func (s *stepCopyImage) copy_progress(ctx context.Context, state multistep.StateBag, dst io.Writer, src io.Reader) error {
	ui := state.Get("ui").(packer.Ui)
	l := utils.NewProgressWriter()
	rdr := io.TeeReader(src, l)

	copyCompleteCh := make(chan error, 1)
	go func() {
		var err error
		_, err = io.Copy(dst, rdr)
		copyCompleteCh <- err
	}()

	progressTicker := time.NewTicker(15 * time.Second)
	defer progressTicker.Stop()

	for {
		select {
		case err := <-copyCompleteCh:

			return err
		case <-progressTicker.C:
			progress := l.Progress()
			if progress.MBytesPerSecond >= 0 {
				ui.Message(fmt.Sprintf("Copy speed: %7.2f MB/s", progress.MBytesPerSecond))
			}
		case <-ctx.Done():
			l.Stop()
			return errors.New("interrupted")
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				ui.Say("Interrupt received. Cancelling copy...")
				l.Stop()
				return errors.New("interrupted")
			}
		}
	}
}

func (s *stepCopyImage) copy(ctx context.Context, state multistep.StateBag, src, dir, filename string) error {

	srcf, err := s.open(src)
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
