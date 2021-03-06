package serv

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wmentor/latency"
	"github.com/wmentor/tt"
	"github.com/wmentor/uniq"
)

type Handler func(c *Context)
type ErrorHandler func(error)
type AuthCheck func(user string, passwd string) bool

type fileHandler struct {
	Filename string
}

func (fh *fileHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, fh.Filename)
}

type node struct {
	name     string
	childs   map[string]*node
	wildCard bool
	fn       Handler
}

type router struct {
	methods           map[string]*node
	redirects         map[string]string
	needUid           bool
	notFoundFunc      Handler
	badRequestFunc    Handler
	internalErrorFunc Handler
	optionsFunc       Handler
	logger            Logger
	longQueryDuration time.Duration
	longQueryHandler  LongQueryHandler
	errorHandler      ErrorHandler
	staticHandlers    map[string]http.Handler
	fileHandlers      map[string]http.Handler
	authCheck         AuthCheck
	tt                *tt.TT
}

func (sr *router) optionsOrNotFound(c *Context) {
	if sr.optionsFunc != nil && c.Method() == "OPTIONS" {
		sr.optionsFunc(c)
	} else {
		sr.notFoundFunc(c)
	}
}

func (r *router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if handler, has := r.fileHandlers[req.URL.Path]; has {
		handler.ServeHTTP(rw, req)
		return
	}

	for pref, handler := range r.staticHandlers {
		if strings.HasPrefix(req.URL.Path, pref) {
			handler.ServeHTTP(rw, req)
			return
		}
	}

	workTime := latency.New()

	ctx := &Context{
		rw:           rw,
		req:          req,
		params:       make(map[string]string),
		errorHandler: r.errorHandler,
		tt:           r.tt,
	}

	defer func() {

		if r.logger != nil {

			ld := &LogData{
				Method:     ctx.Method(),
				Addr:       ctx.RemoteAddr(),
				Auth:       "-",
				RequestURL: ctx.req.RequestURI,
				StatusCode: ctx.statusCode,
				Referer:    ctx.GetHeader("Referer"),
				UserAgent:  ctx.GetHeader("User-Agent"),
				UID:        ctx.Cookie("uid"),
			}

			if user, _, ok := ctx.BasicAuth(); ok {
				ld.Auth = user
			}

			ld.Seconds = workTime.Seconds()

			r.logger(ld)
		}

		if r.longQueryHandler != nil && r.longQueryDuration < workTime.Duration() {
			r.longQueryHandler(workTime.Duration(), ctx)
		}

	}()

	if r.needUid {
		makeUid(rw, req)
	}

	if req.Method == http.MethodGet {
		if dest, has := r.redirects[req.URL.Path]; has {
			ctx.WriteRedirect(dest)
			return
		}
	}

	defer func() {
		if re := recover(); re != nil {
			r.internalErrorFunc(ctx)
			if r.errorHandler != nil {
				r.errorHandler(errors.New(fmt.Sprint(re)))
			}
		}

	}()

	root, has := r.methods[req.Method]
	if !has {
		r.optionsOrNotFound(ctx)
		return
	}

	paths := path2list(req.URL.Path)
	if len(paths) == 0 {
		r.badRequestFunc(ctx)
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

		r.optionsOrNotFound(ctx)
		return
	}

	if root.wildCard {
		params["*"] = tail
	}

	if root.fn != nil {
		ctx.params = params
		root.fn(ctx)
	} else {
		r.optionsOrNotFound(ctx)
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
