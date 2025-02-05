package fasthttp

import (
	"bufio"
	"io"
	"sync"

	"github.com/fr-edy/fasthttp/fasthttputil"
)

// StreamWriter must write data to w.
//
// Usually StreamWriter writes data to w in a loop (aka 'data streaming').
//
// StreamWriter must return immediately if w returns error.
//
// Since the written data is buffered, do not forget calling w.Flush
// when the data must be propagated to reader.
type StreamWriter func(w *bufio.Writer)

// NewStreamReader returns a reader, which replays all the data generated by sw.
//
// The returned reader may be passed to Response.SetBodyStream.
//
// Close must be called on the returned reader after all the required data
// has been read. Otherwise goroutine leak may occur.
//
// See also Response.SetBodyStreamWriter.
func NewStreamReader(sw StreamWriter) io.ReadCloser {
	pc := fasthttputil.NewPipeConns()
	pw := pc.Conn1()
	pr := pc.Conn2()

	var bw *bufio.Writer
	v := streamWriterBufPool.Get()
	if v == nil {
		bw = bufio.NewWriter(pw)
	} else {
		bw = v.(*bufio.Writer)
		bw.Reset(pw)
	}

	go func() {
		sw(bw)
		bw.Flush()
		pw.Close()

		streamWriterBufPool.Put(bw)
	}()

	return pr
}

var streamWriterBufPool sync.Pool
