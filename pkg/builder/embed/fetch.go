package embed

import (
	"compress/gzip"
	"embed"
	"io"
	"io/fs"
)

//go:generate bins/_download_binaries.sh

//go:embed bins
var content embed.FS

type reader struct {
	f fs.File
	g *gzip.Reader
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.g.Read(p)
}

func (r *reader) Close() error {
	r.f.Close()
	return r.g.Close()
}

// try and automatically fetch qemu
func GetEmbededQemu(file string) (io.ReadCloser, error) {
	f, err := content.Open("bins/" + file + ".gz")
	if err != nil {
		return nil, err
	}
	gzf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	return &reader{f: f, g: gzf}, nil
}
