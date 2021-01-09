package serv

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/wmentor/jrpc"
	"github.com/wmentor/tt"
)

type Server struct {
	router *router
	server *http.Server
	jrpc   jrpc.JRPC
}

func New() *Server {

	s := &Server{}

	s.router = &router{
		methods:           make(map[string]*node),
		redirects:         make(map[string]string),
		notFoundFunc:      func(c *Context) { c.StandardError(404) },
		badRequestFunc:    func(c *Context) { c.StandardError(400) },
		internalErrorFunc: func(c *Context) { c.StandardError(500) },
		staticHandlers:    make(map[string]http.Handler),
		fileHandlers:      make(map[string]http.Handler),
		authCheck:         func(login string, passwd string) bool { return false },
		tt:                tt.New(),
	}

	s.jrpc = jrpc.New()

	return s
}

func (s *Server) Start(addr string) error {

	if s.server == nil {
		s.server = &http.Server{Addr: addr, Handler: s.router}
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
	return ErrServerAlreadyStarted
}

func (s *Server) Shutdown() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			if s.router != nil && s.router.errorHandler != nil {
				s.router.errorHandler(err)
			}
		}
		s.server = nil
	}
}

func (s *Server) SetLongQueryHandler(delta time.Duration, fn LongQueryHandler) {
	s.router.longQueryDuration = delta
	s.router.longQueryHandler = fn
}

func (s *Server) SetErrorHandler(fn ErrorHandler) {
	s.router.errorHandler = fn
}

func (s *Server) SetAuthCheck(fn AuthCheck) {
	s.router.authCheck = fn
}

func (s *Server) SetUID(enable bool) {
	s.router.needUid = enable
}

func (s *Server) SetLogger(l Logger) {
	s.router.logger = l
}

func (s *Server) Static(prefix string, dir string) {

	if !strings.HasSuffix(prefix, "/") && prefix != "" && prefix != "/" {
		prefix = prefix + "/"
	}

	handler := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))

	s.router.staticHandlers[prefix] = handler
}

func (s *Server) File(path string, filename string) {
	s.router.fileHandlers[path] = &fileHandler{Filename: filename}
}

func (s *Server) Register(method string, path string, fn Handler) {

	root, has := s.router.methods[method]
	if !has {
		root = &node{childs: make(map[string]*node)}
		s.router.methods[method] = root
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

func (s *Server) RegisterAuth(method string, path string, fn Handler) {

	s.Register(method, path, func(c *Context) {

		if user, login, has := c.BasicAuth(); has {
			if s.router.authCheck(user, login) {
				fn(c)
				return
			}
		}

		c.SetHeader("WWW-Authenticate", `Basic realm="Enter your login and password"`)
		c.WriteHeader(http.StatusUnauthorized)
		c.WriteString("Unauthorized.")
	})
}

func (s *Server) RegMethod(method string, fn interface{}) {
	s.jrpc.RegMethod(method, fn)
}

func (s *Server) RegisterJsonRPC(url string) {

	s.Register("POST", url, func(c *Context) {

		out := bytes.NewBuffer(nil)

		if err := s.jrpc.Process(c.Body(), out); err == nil {
			c.SetContentType("application/json; charset=utf-8")
			c.WriteHeader(200)
			c.Write(out.Bytes())
		} else {
			c.StandardError(400)
		}

	})

}

func (s *Server) LoadTemplates(dir string) {
	s.router.tt = tt.New(dir)
}
