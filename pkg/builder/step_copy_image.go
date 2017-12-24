package builder

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	filetype "gopkg.in/h2non/filetype.v1"
	"gopkg.in/h2non/filetype.v1/matchers"
)

type stepCopyImage struct {
	FromKey, ResultKey string
	ui                 packer.Ui
}

func (s *stepCopyImage) Run(state multistep.StateBag) multistep.StepAction {
	fromFile := state.Get(s.FromKey).(string)
	config := state.Get("config").(*Config)
	s.ui = state.Get("ui").(packer.Ui)
	s.ui.Say("Copying source image.")

	dstfile := filepath.Join(config.OutputDir, "image")
	err := s.copy(fromFile, config.OutputDir, "image")
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

func (s *stepCopyImage) copy(src, dir, filename string) error {

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

	_, err = io.Copy(dstf, srcf)

	if err != nil {
		return err
	}

	return nil
}
