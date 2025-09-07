package requests

import (
	"bytes"
	"errors"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"strings"
)

const crlf = "\r\n"
const bufferSize = 8

type requestParsingState int

const (
	initialized requestParsingState = iota
	parsingHeader
	done
)

type Request struct {
	RequestLine         RequestLine
	Headers             headers.Headers
	requestParsingState requestParsingState
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
		requestParsingState: initialized,
		Headers:             headers.NewHeaders(),
	}

	for currRequest.requestParsingState != done {

		if len(buf) <= readToIndex {
			newSlice := make([]byte, len(buf)*2)
			copy(newSlice, buf)
			buf = newSlice
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if currRequest.requestParsingState != done {
					return nil, errors.New("Incomplete Request")
				}
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

func (r *Request) singleParse(data []byte) (int, error) {
	switch r.requestParsingState {
	case initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.requestParsingState = parsingHeader
		return n, nil

	case parsingHeader:
		n, finished, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if finished {
			r.requestParsingState = done
			return n, nil
		}
		return n, nil

	case done:
		return 0, errors.New("Error: Trying to read data in a done state")
	default:
		return 0, errors.New("Unknown")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.requestParsingState != done {
		n, err := r.singleParse(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}
