package request

import (
	"errors"
	"io"
)

type rewindReader struct {
	io.ReadSeeker
	startPos int64
}

func newRewindReader(body io.Reader) (*rewindReader, error) {
	sr, ok := body.(io.ReadSeeker)
	if !ok {
		return nil, errors.New("body doesn't implement io.Seeker interface")
	}
	startPos, err := sr.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	return &rewindReader{ReadSeeker: sr, startPos: startPos}, nil
}

func (r *rewindReader) Rewind() error {
	_, err := r.ReadSeeker.Seek(r.startPos, io.SeekStart)
	return err
}
