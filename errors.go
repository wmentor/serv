package serv

import (
	"errors"
	"net/http"
)

var (
	errorCodes map[int][]byte

	errorInvalidRequestMethod error = errors.New("invalid request method")
)

func init() {

	errorCodes = map[int][]byte{
		400: []byte("400 Bad Request"),
		401: []byte("401 Unauthorized"),
		403: []byte("403 Forbidden"),
		404: []byte("404 Status Not Found"),
		405: []byte("405 Method Not Allowed"),
		409: []byte("409 Conflict"),
		429: []byte("429 Too Many Requests"),
		500: []byte("500 Internal Server Error"),
	}

}

func SendError(rw http.ResponseWriter, code int) {

	_, h := errorCodes[code]

	if !h {
		code = 500
	}

	m, _ := errorCodes[code]

	rw.WriteHeader(code)
	rw.Write([]byte(m))
}
