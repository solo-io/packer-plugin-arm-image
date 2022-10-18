package embed

import (
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"runtime"
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

	// for now, we only embed linux amd64 qemu
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		return nil, fmt.Errorf("currently, embedded qemu is only available for linux amd64. please download qemu-user-static manually")
	}

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
