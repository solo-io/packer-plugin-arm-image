package builder

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

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

func (s *stepCopyImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	fromFile := state.Get(s.FromKey).(string)
	config := state.Get("config").(*Config)
	s.ui = state.Get("ui").(packer.Ui)
	s.ui.Say("Copying source image.")

	dstfile := filepath.Join(config.OutputDir, "image")
	err := s.copy(state, fromFile, config.OutputDir, "image")
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

func (s *stepCopyImage) openxz(f *os.File) (io.ReadCloser, error) {
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

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

type ProgressWriter struct {
	done      int32
	totalData uint64

	lastProgressData uint64
	lastProgressTime time.Time
}

func NewProgressWriter() *ProgressWriter {
	return &ProgressWriter{
		lastProgressTime: time.Now(),
	}
}
func (pw *ProgressWriter) Write(data []byte) (int, error) {
	if atomic.LoadInt32(&pw.done) != 0 {
		return 0, errors.New("copy interrupted")
	}
	atomic.AddUint64(&pw.totalData, uint64(len(data)))
	return len(data), nil
}

func (pw *ProgressWriter) Progress() float64 {
	currentData := atomic.LoadUint64(&pw.totalData)
	now := time.Now()
	deltat := now.Sub(pw.lastProgressTime)
	deltadata := currentData - pw.lastProgressData

	pw.lastProgressData = currentData
	pw.lastProgressTime = now
	// TODO: is this the right way to measure? maybe change 1e6 to float64(1 << 20)?
	return (float64(deltadata) / 1e6) / deltat.Seconds()
}

func (pw *ProgressWriter) Stop() {
	atomic.StoreInt32(&pw.done, 1)
}

func (s *stepCopyImage) copy_progress(state multistep.StateBag, dst io.Writer, src io.Reader) error {
	ui := state.Get("ui").(packer.Ui)
	l := NewProgressWriter()
	rdr := io.TeeReader(src, l)

	copyCompleteCh := make(chan error, 1)
	go func() {
		var err error
		_, err = io.Copy(dst, rdr)
		copyCompleteCh <- err
	}()

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

	for {
		select {
		case err := <-copyCompleteCh:

			return err
		case <-progressTicker.C:
			progress := l.Progress()
			if progress >= 0 {
				ui.Message(fmt.Sprintf("Copy speed: %f MB/s", progress))
			}
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				ui.Say("Interrupt received. Cancelling copy...")
				l.Stop()
				return errors.New("interrupted")
			}
		}
	}
}

func (s *stepCopyImage) copy(state multistep.StateBag, src, dir, filename string) error {

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

	err = s.copy_progress(state, dstf, srcf)

	if err != nil {
		return err
	}

	return nil
}
