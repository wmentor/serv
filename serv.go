package serv

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Context struct {
	RW      http.ResponseWriter
	Request *http.Request
	params  Params
}

type Handler func(c *Context)

type node struct {
	name     string
	childs   map[string]*node
	wildCard bool
	fn       Handler
}

type serv struct {
	methods           map[string]*node
	redirects         map[string]string
	notFoundFunc      http.HandlerFunc
	badRequestFunc    http.HandlerFunc
	internalErrorFunc http.HandlerFunc
	optionsFunc       http.HandlerFunc
}

var (
	rt         *serv
	errorCodes map[int]string
)

func init() {

	rt = &serv{
		methods:           make(map[string]*node),
		redirects:         make(map[string]string),
		notFoundFunc:      func(rw http.ResponseWriter, req *http.Request) { SendErrorCode(rw, 404) },
		badRequestFunc:    func(rw http.ResponseWriter, req *http.Request) { SendErrorCode(rw, 400) },
		internalErrorFunc: func(rw http.ResponseWriter, req *http.Request) { SendErrorCode(rw, 500) },
	}

	errorCodes = map[int]string{
		400: "400 Bad Request",
		401: "401 Unauthorized",
		403: "403 Forbidden",
		404: "404 Status Not Found",
		405: "405 Method Not Allowed",
		409: "409 Conflict",
		429: "429 Too Many Requests",
		500: "500 Internal Server Error",
	}
}

func SendErrorCode(rw http.ResponseWriter, code int) {

	_, h := errorCodes[code]

	if !h {
		code = 500
	}

	m, _ := errorCodes[code]

	rw.WriteHeader(code)
	fmt.Fprint(rw, m)
}

func Register(method string, path string, fn Handler) {

	root, has := rt.methods[method]
	if !has {
		root = &node{childs: make(map[string]*node)}
		rt.methods[method] = root
	}

	list := path2list(path)
	if len(list) == 0 {
		return
	}

	for _, item := range list {

		if item[0] == ':' {
			n, h := root.childs[""]
			if !h {
				name := item
				if len(name) > 1 {
					name = item[1:]
				} else {
					name = ""
				}
				n = &node{name: name, wildCard: false, childs: make(map[string]*node)}
			}
			root.childs[""] = n
			root = n
		} else if item == "*" {
			_, h := root.childs[""]
			if !h {
				root.childs[""] = &node{name: "*", fn: fn, wildCard: true}
			}
			return
		} else {

			n, h := root.childs[item]
			if !h {
				n = &node{name: "", wildCard: false, childs: make(map[string]*node)}
			}
			root.childs[item] = n
			root = n
		}
	}

	root.fn = fn
}

func path2list(path string) []string {

	if len(path) < 1 || path[0] != '/' {
		return nil
	}

	list := make([]string, 1, 32)

	list[0] = "/"

	for _, v := range strings.Split(path[1:], "/") {
		if v != "" {
			if uri, e := url.PathUnescape(v); e == nil {
				list = append(list, uri)
				if uri == "*" {
					return list
				}
			} else {
				return nil
			}

		}
	}

	return list
}
