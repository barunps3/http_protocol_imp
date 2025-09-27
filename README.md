# Implement HTTP/1.1 basic Webserver from scratch in Go
In this small project I try to implement Webserver from scratch without using `http` go package.
This Webserver, can do the following:
- Listen for requests on port `:42069` for HTTP requests
- Parses bytes of data received on the above port to create `requests.Requests` object
- `requests` object is created based on [RFC 9112 Message Format](https://datatracker.ietf.org/doc/html/rfc9112#name-message-format)
- Currently on Responds to `GET` Requests, however can be extended to handle other `HTTP Methods`
- Response body follows the same [RFC 9112 Message Format](https://datatracker.ietf.org/doc/html/rfc9112#name-message-format) 
- Response Headers always include `Content-Length` (unless  Chunked-Encoding) and `Connection: Close` header for `curl` to close the connection safely
- Included test cases to check parsing of request and response based on message format
- Capable of sending response body in `Chunked-Encoding` format as well
- Server shutsdown with interrupt (Ctrl + C)`SIGINT`

## How to execute
Run the following command in a terminal window
```
go run cmd/httpserver/main.go
```
In another window, test with command
```
curl -v localhost:42069/yourproblem
```


### Chunked Encoding
To test chunked encoding, run the following command.
It returns the 100 JSON objects returned from https://httpbin.org/stream/100 in maxChunkSize of 1024 bytes
```
curl --raw -v localhost:42069/httpbin/stream/100
```
Response returned in chunked encoding are in the following [RFC 9112 Transfer-Encoding Format](https://datatracker.ietf.org/doc/html/rfc9112#field.transfer-encoding):
```
HTTP/1.1 200 OK
Content-Type: text/plain
Transfer-Encoding: chunked

<n>\r\n
<data of length n>\r\n
<n>\r\n
<data of length n>\r\n
<n>\r\n
<data of length n>\r\n
<n>\r\n
<data of length n>\r\n
... repeat ...
0\r\n
\r\n
```

### Video Content Type
Save a video file in folder `assets/`
And in your browser type enter the address `http://localhost:42069/video`
As of now, it can handle small video files only!!
