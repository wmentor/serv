package serv

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wmentor/jrpc"
	"github.com/wmentor/uniq"
)

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
	needUid           bool
	notFoundFunc      http.HandlerFunc
	badRequestFunc    http.HandlerFunc
	internalErrorFunc http.HandlerFunc
	optionsFunc       http.HandlerFunc
}

var (
	rt *serv
)

func init() {

	rt = &serv{
		methods:           make(map[string]*node),
		redirects:         make(map[string]string),
		notFoundFunc:      func(rw http.ResponseWriter, req *http.Request) { SendError(rw, 404) },
		badRequestFunc:    func(rw http.ResponseWriter, req *http.Request) { SendError(rw, 400) },
		internalErrorFunc: func(rw http.ResponseWriter, req *http.Request) { SendError(rw, 500) },
	}
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

func RegMethod(method string, fn interface{}) {
	jrpc.RegMethod(method, fn)
}

func RegisterJsonRPC(url string) {

	Register("POST", url, func(c *Context) {

		data, err := ioutil.ReadAll(c.Body())
		if err != nil {
			SendError(c.rw, 500)
		}

		if data, err = jrpc.Handle(data); err != nil {
			SendError(c.rw, 400)
		}

		c.SetContentType("application/json; charset=utf-8")
		c.WriteHeader(200)
		c.Write(data)

	})

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

func (r *serv) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if r.needUid {
		makeUid(rw, req)
	}

	if req.Method == http.MethodGet {

		if dest, has := r.redirects[req.URL.Path]; has {
			http.Redirect(rw, req, dest, 302)
			return
		}
	}

	defer func() {

		if re := recover(); re != nil {
			r.internalErrorFunc(rw, req)
		}

	}()

	if req.Method == http.MethodOptions && r.optionsFunc != nil {
		r.optionsFunc(rw, req)
		return
	}

	root, has := r.methods[req.Method]
	if !has {
		r.notFoundFunc(rw, req)
		return
	}

	paths := path2list(req.URL.Path)
	if len(paths) == 0 {
		r.badRequestFunc(rw, req)
		return
	}

	tail := ""
	params := make(map[string]string)

	for _, item := range paths {

		if root.wildCard {
			tail += "/" + item
			continue
		}

		if n, h := root.childs[""]; h {
			root = n
			if root.wildCard {
				tail = "/" + item
			} else {
				params[root.name] = item
			}
			continue
		}

		if n, h := root.childs[item]; h {
			root = n
			continue
		}

		r.notFoundFunc(rw, req)
		return
	}

	if root.wildCard {
		params["*"] = tail
	}

	if root.fn != nil {

		ctx := &Context{
			rw:     rw,
			req:    req,
			params: params,
		}

		root.fn(ctx)
	} else {
		r.notFoundFunc(rw, req)
	}
}

func makeUid(rw http.ResponseWriter, req *http.Request) {

	c, err := req.Cookie("uid")

	v := ""

	if err != nil {
		v = uniq.New()
	} else {
		v = c.Value
	}

	cookie := &http.Cookie{
		Name:     "uid",
		Value:    v,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(time.Now().Unix()+86400*366, 0),
	}

	req.AddCookie(cookie)

	http.SetCookie(rw, cookie)
}

func Start(addr string) error {
	return http.ListenAndServe(addr, rt)
}

func SetUID(enable bool) {
	rt.needUid = enable
}
