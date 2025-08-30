package requests

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

const crlf = "\r\n"
const bufferSize = 8

type parserState int

const (
	initialized parserState = iota
	done
)

type Request struct {
	RequestLine RequestLine
	parserState parserState
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.parserState {
	case initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.parserState = done
		return n, nil
	case done:
		return 0, errors.New("Error: Trying to read data in a done state")
	default:
		return 0, errors.New("Unknown")
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func requestLineFromString(reqLineStr string) (*RequestLine, error) {
	requestLine := &RequestLine{}
	reqLine := strings.Split(reqLineStr, " ")
	if len(reqLine) != 3 {
		return nil, errors.New("HTTP Request Line not complete")
	}
	methodStr := reqLine[0]
	matched, err := regexp.MatchString(`^[A-Z]+$`, methodStr)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.New("Method does not match any of the HTTP/1.1 method")
	}

	requestTarget := reqLine[1]
	httpNameVersion := reqLine[2]

	if httpNameVersion != "HTTP/1.1" {
		return nil, errors.New("HTTP version not supported")
	}
	httpVersion := strings.Split(httpNameVersion, "/")[1]

	// Fill in the requestline
	requestLine.Method = methodStr
	requestLine.RequestTarget = requestTarget
	requestLine.HttpVersion = httpVersion

	return requestLine, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// need more data
		return nil, 0, nil
	}
	reqLineStr := string(data[:idx])
	requestLine, err := requestLineFromString(reqLineStr)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0

	currRequest := &Request{
		parserState: initialized,
	}

	for currRequest.parserState != done {

		if len(buf) <= readToIndex {
			newSlice := make([]byte, len(buf)*2)
			copy(newSlice, buf)
			buf = newSlice
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				currRequest.parserState = done
				break
			}
			return nil, err
		}

		readToIndex = readToIndex + n
		parsedByteCount, err := currRequest.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[parsedByteCount:])
		readToIndex = readToIndex - parsedByteCount
	}

	return currRequest, nil
}
