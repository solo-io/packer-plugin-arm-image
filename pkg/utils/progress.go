package utils

import (
	"errors"
	"sync/atomic"
	"time"
)

type ProgressWriter struct {
	done      int32
	totalData uint64
	fileSize  uint64

	lastProgressData uint64
	lastProgressTime time.Time
}

func NewProgressWriter() *ProgressWriter {
	return &ProgressWriter{
		lastProgressTime: time.Now(),
	}
}

func NewProgressWriterWithSize(fileSize uint64) *ProgressWriter {
	pw := NewProgressWriter()
	pw.fileSize = fileSize
	return pw
}

func (pw *ProgressWriter) Write(data []byte) (int, error) {
	if atomic.LoadInt32(&pw.done) != 0 {
		return 0, errors.New("copy interrupted")
	}
	atomic.AddUint64(&pw.totalData, uint64(len(data)))
	return len(data), nil
}

type Progress struct {
	MBytesPerSecond float64
	PercentDone     float64
}

func (pw *ProgressWriter) TotalData() uint64 {
	return atomic.LoadUint64(&pw.totalData)
}

func (pw *ProgressWriter) Progress() Progress {
	currentData := pw.TotalData()
	now := time.Now()
	deltat := now.Sub(pw.lastProgressTime)
	deltadata := currentData - pw.lastProgressData

	pw.lastProgressData = currentData
	pw.lastProgressTime = now

	progress := Progress{
		MBytesPerSecond: (float64(deltadata) / 1e6) / deltat.Seconds(),
		PercentDone:     -1,
	}
	if pw.fileSize > 0 {
		// TODO: is this the right way to measure? maybe change 1e6 to float64(1 << 20)?
		progress.PercentDone = 100.0 * float64(currentData) / float64(pw.fileSize)
	}

	return progress
}

func (pw *ProgressWriter) Stop() {
	atomic.StoreInt32(&pw.done, 1)
}
