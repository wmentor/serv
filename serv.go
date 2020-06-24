package serv

import (
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
	methods   map[string]*node
	redirects map[string]string
}

var (
	rt *serv
)

func init() {

	rt = &serv{
		methods:   make(map[string]*node),
		redirects: make(map[string]string),
	}
}

func Register(method string, path string, fn Handler) {

	if rt.methods[method] == nil {
		rt.methods[method] = &node{
			childs: make(map[string]*node),
		}
	}

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
			} else {
				return nil
			}

		}
	}

	return list
}
