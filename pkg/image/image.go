package image

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	filetype "gopkg.in/h2non/filetype.v1"
	"gopkg.in/h2non/filetype.v1/matchers"

	"compress/bzip2"
	"compress/gzip"

	"github.com/ulikunitz/xz"
)

type nilUi struct{}

type imageOpener struct {
	ui packer.Ui
}

func (*nilUi) Ask(string) (string, error) {
	return "", errors.New("no ui available")
}
func (*nilUi) Say(string) {

}
func (*nilUi) Message(string) {

}
func (*nilUi) Error(string) {

}
func (*nilUi) Machine(string, ...string) {

}

func (*nilUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	return stream
}

func NewImageOpener(ui packer.Ui) ImageOpener {
	if ui == nil {
		ui = &nilUi{}
	}
	return &imageOpener{ui: ui}
}

type fileImage struct {
	io.ReadCloser
	size uint64
}

func (f *fileImage) SizeEstimate() uint64 { return f.size }
func openImage(file *os.File) (Image, error) {

	finfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fsize := finfo.Size()

	ret := fileImage{ReadCloser: file, size: uint64(fsize)}
	return &ret, nil

}

func (s *imageOpener) Open(fpath string) (Image, error) {
	t, _ := filetype.MatchFile(fpath)

	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	switch t {
	case matchers.TypeZip:
		s.ui.Say("Image is a zip file.")
		return s.openzip(f)
	case matchers.TypeXz:
		s.ui.Say("Image is a xz file.")
		return s.openxz(f)
	case matchers.TypeGz:
		s.ui.Say("Image is a gzip file.")
		return s.opengzip(f)
	case matchers.TypeBz2:
		s.ui.Say("Image is a gzip file.")
		return s.openbzip(f)
	default:
		return openImage(f)
	}

}

func (s *imageOpener) openzip(f *os.File) (Image, error) {
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
	mc := &multiCloser{zippedfileReader, []io.Closer{zippedfileReader, f}, nil, zippedfile.UncompressedSize64}
	f = nil

	return mc, nil
}

func (s *imageOpener) openxz(f *os.File) (Image, error) {
	return uncompress(f, "xzcat", func(r io.Reader) (io.Reader, error) { r2, e := xz.NewReader(r); return r2, e })
}

func (s *imageOpener) opengzip(f *os.File) (Image, error) {

	return uncompress(f, "zcat", func(r io.Reader) (io.Reader, error) { r2, e := gzip.NewReader(r); return r2, e })
}

func (s *imageOpener) openbzip(f *os.File) (Image, error) {

	return uncompress(f, "bzcat", func(r io.Reader) (io.Reader, error) { r2 := bzip2.NewReader(r); return r2, nil })
}

func uncompress(f *os.File, fastcmd string, slowNewReader func(r io.Reader) (io.Reader, error)) (Image, error) {
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	// check if available:
	if exec.Command("which", fastcmd).Run() == nil {
		ret, err := xzFastlane(fastcmd, f)
		if err == nil {
			f = nil
			return ret, err
		}
	}
	// slow lane here
	r, err := slowNewReader(f)
	if err != nil {
		return nil, err
	}

	//transfer ownership
	mc := &multiCloser{r, []io.Closer{f}, nil, 0}
	f = nil

	return mc, nil
}

func xzFastlane(cmd string, f *os.File) (Image, error) {

	xzcat := exec.Command(cmd)

	// fast path, use xzcat
	xzcat.Stdin = f
	r, err := xzcat.StdoutPipe()

	if err != nil {
		return nil, err
	}
	if err := xzcat.Start(); err != nil {
		return nil, err
	}

	// use mc for size estimate
	mc := &multiCloser{r, []io.Closer{r}, xzcat, 0}

	return mc, nil

}

type multiCloser struct {
	io.Reader
	c []io.Closer
	// As the reader that is stored in c is the StdoutPipe of an exec.Cmd
	// instance, we must ensure the command isn't closed prior to the reader
	// being closed otherwise we end up with and error like:
	// 'read |0: file already closed'.
	// See also https://pkg.go.dev/os/exec#Cmd.StdoutPipe
	command *exec.Cmd

	sizeEstimate uint64
}

func (n *multiCloser) Close() error {
	for _, c := range n.c {
		c.Close()
	}

	var err error
	// Wait for the readers (including StdoutPipe) to be closed
	// and then initiate the command shutdown if we have a command.
	if n.command != nil {
		err = n.command.Wait()
	}
	return err
}

func (f *multiCloser) SizeEstimate() uint64 { return f.sizeEstimate }
