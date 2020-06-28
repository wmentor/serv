package serv

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type Context struct {
	rw     http.ResponseWriter
	req    *http.Request
	params Params
	qw     Query
}

func (c *Context) Write(data []byte) {
	c.rw.Write(data)
}

func (c *Context) WriteJson(v interface{}) {

	encoder := json.NewEncoder(c.rw)
	if err := encoder.Encode(v); err != nil {
		SendError(c.rw, 500)
	}
}

func (c *Context) WriteRedirect(dest string) {
	http.Redirect(c.rw, c.req, dest, 302)
}

func (c *Context) WriteHeader(code int) {
	c.rw.WriteHeader(code)
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
