package image

import (
	"io"
)

type ImageOpener interface {
	Open(filename string) (Image, error)
}

type ImageFile interface {
	Open() (Image, error)
	Name()
}

type Image interface {
	io.ReadCloser
	SizeEstimate() uint64
}
