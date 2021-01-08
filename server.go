package serv

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	router *router
	server *http.Server
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
	}

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
			if rt != nil && s.router.errorHandler != nil {
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
