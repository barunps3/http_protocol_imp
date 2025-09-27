package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/requests"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *requests.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler200(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerHttpBin(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
		return
	}
}

func handlerHttpBin(w *response.Writer, req *requests.Request) {
	forwardReqTarget := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	forwardUrl := fmt.Sprintf("https://httpbin.org/%s", forwardReqTarget)
	fmt.Println("Proxying to", forwardUrl)
	resp, err := http.Get(forwardUrl)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	fullBody := make([]byte, 0)

	const maxChunkSize = 1024
	buf := make([]byte, maxChunkSize)
	for {
		n, err := resp.Body.Read(buf)
		fmt.Println("Read: ", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body: ", err)
				break
			}
			fullBody = append(fullBody, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response: ", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done: ", err)
	}
	trailers := headers.NewHeaders()
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Override("X-Content-SHA256", sha256)
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers: ", err)
	}
	fmt.Println("Wrote Trailer")
}

func handlerVideo(w *response.Writer, req *requests.Request) {
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}
	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(len(data))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(data)
}

func handler400(w *response.Writer, _ *requests.Request) {
	w.WriteStatusLine(response.BadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *requests.Request) {
	w.WriteStatusLine(response.InternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *requests.Request) {
	w.WriteStatusLine(response.OK)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}
