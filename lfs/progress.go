package lfs

import (
	"io"
	"time"
)

func ProgressReader(r io.Reader, responder func(Response), oid string, size int64) io.Reader {
	return &progressReader{
		inner:     r,
		oid:       oid,
		size:      size,
		responder: responder,
	}
}

type progressReader struct {
	inner     io.Reader
	oid       string
	size      int64
	responder func(Response)

	bytesSoFar     int64
	bytesSinceLast int64
	lastEvent      time.Time
}

func (r *progressReader) Read(b []byte) (int, error) {
	n, err := r.inner.Read(b)

	r.bytesSoFar += int64(n)
	r.bytesSinceLast += int64(n)
	if time.Since(r.lastEvent) > time.Second || err == io.EOF {
		r.responder(TransferProgress(r.oid, r.bytesSoFar, r.bytesSinceLast))
		r.lastEvent = time.Now()
		r.bytesSinceLast = 0
	}

	return n, err
}
