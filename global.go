package serv

import (
	"time"
)

var (
	server *Server
)

func init() {
	server = New()
}

func Start(addr string) error {
	return server.Start(addr)
}

func Shutdown() {
	server.Shutdown()
}

func SetLongQueryHandler(delta time.Duration, fn LongQueryHandler) {
	server.SetLongQueryHandler(delta, fn)
}

func SetErrorHandler(fn ErrorHandler) {
	server.SetErrorHandler(fn)
}

func SetAuthCheck(fn AuthCheck) {
	server.SetAuthCheck(fn)
}

func SetUID(enable bool) {
	server.SetUID(enable)
}

func SetLogger(l Logger) {
	server.SetLogger(l)
}

func Static(prefix string, dir string) {
	server.Static(prefix, dir)
}

func File(path string, filename string) {
	server.File(path, filename)
}

func Register(method string, path string, fn Handler) {
	server.Register(method, path, fn)
}

func RegisterAuth(method string, path string, fn Handler) {
	server.RegisterAuth(method, path, fn)
}

func RegMethod(method string, fn interface{}) {
	server.RegMethod(method, fn)
}

func RegisterJsonRPC(url string) {
	server.RegisterJsonRPC(url)
}

func LoadTemplates(dir string) {
	server.LoadTemplates(dir)
}
