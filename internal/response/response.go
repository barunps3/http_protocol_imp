package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

type writerState int

const (
	writerStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
)

type Writer struct {
	writer      io.Writer
	writerState writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: writerStatusLine,
		writer:      w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateHeaders }()
	reasonPhrase := "HTTP/1.1 "
	switch statusCode {
	case OK:
		reasonPhrase += fmt.Sprintf("%d OK", OK)
	case BadRequest:
		reasonPhrase += fmt.Sprintf("%d Bad Request", BadRequest)
	case InternalServerError:
		reasonPhrase += fmt.Sprintf("%d Internal Server Error", InternalServerError)
	default:
		reasonPhrase += fmt.Sprintf("%d ", statusCode)
	}
	reasonPhrase += "\r\n"
	_, err := w.writer.Write([]byte(reasonPhrase))
	return err

}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()
	headersString := ""
	for key, value := range headers {
		headersString += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	headersString += "\r\n"
	_, err := w.writer.Write([]byte(headersString))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	return w.writer.Write(p)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}

	hexUpperStr := fmt.Sprintf("%X\r\n", len(p))
	chunkBody := append([]byte(hexUpperStr), p...)
	endChars := []byte("\r\n")
	chunkBody = append(chunkBody, endChars...)
	return w.writer.Write(chunkBody)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	w.writerState = writerStateTrailers
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("cannot write body in state %d", w.writerState)
	}

	headersString := ""
	for key, value := range h {
		headersString += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	headersString += "\r\n"
	_, err := w.writer.Write([]byte(headersString))
	return err
}
