package serv

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/wmentor/tt"
)

type Context struct {
	rw           http.ResponseWriter
	req          *http.Request
	params       Params
	qw           Query
	statusCode   int
	errorHandler ErrorHandler
	tt           *tt.TT
}

func (c *Context) StandardError(code int) {
	_, h := errorCodes[code]

	if !h {
		code = 500
	}

	m, _ := errorCodes[code]

	c.SetContentType("text/plain; charset=utf-8")
	c.WriteHeader(code)
	c.Write([]byte(m))
}

func (c *Context) Write(data []byte) {
	c.rw.Write(data)
}

func (c *Context) WriteJson(v interface{}) {

	encoder := json.NewEncoder(c.rw)
	if err := encoder.Encode(v); err != nil {
		c.StandardError(500)
	}
}

func (c *Context) WriteString(txt string) {
	c.rw.Write([]byte(txt))
}

func (c *Context) WriteRedirect(dest string) {
	if c.statusCode == 0 {
		c.statusCode = 302
		http.Redirect(c.rw, c.req, dest, 302)
	}
}

func (c *Context) WritePermanentRedirect(dest string) {
	if c.statusCode == 0 {
		c.statusCode = 301
		http.Redirect(c.rw, c.req, dest, 301)
	}
}

func (c *Context) WriteHeader(code int) {
	if c.statusCode == 0 {
		c.statusCode = code
		c.rw.WriteHeader(code)
	}
}

func (c *Context) SetHeader(key, value string) {
	c.rw.Header().Add(key, value)
}

func (c *Context) GetHeader(key string) string {
	return c.req.Header.Get(key)
}

func (c *Context) GetContentType() string {
	return c.GetHeader("Content-Type")
}

func (c *Context) SetContentType(value string) {
	c.SetHeader("Content-Type", value)
}

func (c *Context) Param(name string) string {
	return c.params.GetString(name)
}

func (c *Context) ParamInt(name string) int {
	return c.params.GetInt(name)
}

func (c *Context) ParamInt64(name string) int64 {
	return c.params.GetInt64(name)
}

func (c *Context) ParamBool(name string) bool {
	return c.params.GetBool(name)
}

func (c *Context) ParamFloat(name string) float64 {
	return c.params.GetFloat(name)
}

func (c *Context) query() Query {
	if c.qw == nil {
		c.qw = newQuery(c.req)
	}
	return c.qw
}

func (c *Context) Query(name string) string {
	return c.query().GetString(name)
}

func (c *Context) QueryInt(name string) int {
	return c.query().GetInt(name)
}

func (c *Context) QueryInt64(name string) int64 {
	return c.query().GetInt64(name)
}

func (c *Context) QueryBool(name string) bool {
	return c.query().GetBool(name)
}

func (c *Context) QueryFloat(name string) float64 {
	return c.query().GetFloat(name)
}

func (c *Context) HasQueryParam(name string) bool {
	return c.query().Has(name)
}

func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	f, fh, err := c.req.FormFile(name)
	return f, fh, err
}

func (c *Context) FormValue(name string) string {
	return c.req.FormValue(name)
}

func (c *Context) FormValueInt(name string) int {
	if res, err := strconv.Atoi(c.req.FormValue(name)); err == nil {
		return res
	}
	return 0
}

func (c *Context) FormValueInt64(name string) int64 {
	if res, err := strconv.ParseInt(c.req.FormValue(name), 10, 64); err == nil {
		return res
	}
	return 0
}

func (c *Context) FormValueBool(name string) bool {
	if res, err := strconv.ParseBool(c.req.FormValue(name)); err == nil {
		return res
	}
	return false
}

func (c *Context) FormValueFloat(name string) float64 {
	if res, err := strconv.ParseFloat(c.req.FormValue(name), 64); err == nil {
		return res
	}
	return 0
}

func (c *Context) Method() string {
	return c.req.Method
}

func (c *Context) Body() io.ReadCloser {
	return c.req.Body
}

func (c *Context) BodyJson(res interface{}) error {

	if c.req.Body == nil {
		return errors.New("empty body")
	}

	m := c.Method()

	if m == "POST" || m == "PUT" {
		decoder := json.NewDecoder(c.Body())
		return decoder.Decode(res)
	}

	return errorInvalidRequestMethod
}

func (c *Context) BasicAuth() (string, string, bool) {
	return c.req.BasicAuth()
}

func (c *Context) Cookie(name string) string {
	cookie, err := c.req.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	if cookie != nil {
		http.SetCookie(c.rw, cookie)
	}
}

func (c *Context) Render(tmpl string, vars map[string]interface{}) {

	v := c.tt.MakeVars()

	for k, val := range vars {
		v.Set(k, val)
	}

	if res, err := c.tt.Render(tmpl, v); err == nil {
		c.Write(res)
	} else {
		if c.errorHandler != nil {
			c.errorHandler(err)
		}
	}
}

func (c *Context) RenderStr(tmpl string, vars map[string]interface{}) {

	v := c.tt.MakeVars()

	for k, val := range vars {
		v.Set(k, val)
	}

	if res, err := c.tt.RenderString(tmpl, v); err == nil {
		c.Write(res)
	} else {
		if c.errorHandler != nil {
			c.errorHandler(err)
		}
	}
}

func (c *Context) RemoteAddr() string {

	if v := c.GetHeader("X-Forwarded-For"); len(v) > 0 {
		addrs := strings.Split(v, ",")
		if addrs != nil && len(addrs) > 0 {
			ip := strings.TrimSpace(addrs[0])
			if len(ip) < 20 {
				return ip
			}
		}
	}

	if ip := c.GetHeader("X-Real-Ip"); len(ip) > 0 && len(ip) < 20 {
		return ip
	}

	if ip, _, err := net.SplitHostPort(c.req.RemoteAddr); err == nil {
		if len(ip) < 20 {
			return (ip)
		}
	}

	return ""
}

func (c *Context) RequestPath() string {
	return c.req.URL.Path
}

func (c *Context) RequestURI() string {
	return c.req.RequestURI
}
