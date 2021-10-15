package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/solo-io/packer-plugin-arm-image/pkg/image"
)

type copyResult struct {
	n   int64
	err error
}

func CopyWithProgress(ctx context.Context, ui packer.Ui, dst io.Writer, src io.Reader) (int64, error) {
	var l *ProgressWriter
	if image, ok := src.(image.Image); ok {
		l = NewProgressWriterWithSize(image.SizeEstimate())
	} else {
		l = NewProgressWriter()
	}
	// this will cancel the Copy if still in progress when we leave the function
	defer l.Stop()
	rdr := io.TeeReader(src, l)

	copyCompleteCh := make(chan copyResult, 1)
	go func() {
		defer close(copyCompleteCh)
		n, err := io.Copy(dst, rdr)
		copyCompleteCh <- copyResult{n: n, err: err}
	}()

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

	for {
		select {
		case res := <-copyCompleteCh:
			return res.n, res.err
		case <-progressTicker.C:
			progress := l.Progress()
			if progress.MBytesPerSecond >= 0 {
				ui.Message(fmt.Sprintf("Speed: %7.2f MB/s", progress.MBytesPerSecond))
			}
			if progress.PercentDone > 0 {
				ui.Message(fmt.Sprintf("Progress: %3.2f%%", progress.PercentDone))
			}
		case <-ctx.Done():
			return int64(l.TotalData()), errors.New("interrupted")
		}
	}
}
