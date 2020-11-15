package serv

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wmentor/jrpc"
)

func TestPath(t *testing.T) {

	tF := func(path string, wait []string) {
		res := path2list(path)
		if strings.Join(res, "#") != strings.Join(wait, "#") {
			t.Fatalf("path2list faild for: %s", path)
		}
	}

	tF("", nil)
	tF("123", nil)
	tF("/", []string{"/"})
	tF("//", []string{"/"})
	tF("/test", []string{"/", "test"})
	tF("/test/", []string{"/", "test"})
	tF("/test+test/", []string{"/", "test+test"})
	tF("/hello/world", []string{"/", "hello", "world"})
	tF("/hello/:login/", []string{"/", "hello", ":login"})
	tF("/posts/*", []string{"/", "posts", "*"})
}

func TestServ(t *testing.T) {

	Register("GET", "/", func(c *Context) {
		c.SetContentType("text/plain; charset=utf-8")
		c.WriteHeader(200)
		c.WriteString("Hello!")
	})

	Register("GET", "/user/:user", func(c *Context) {
		c.SetContentType("text/plain; charset=utf-8")
		c.WriteHeader(200)
		c.WriteString("Hello, " + c.Param("user") + "!")
	})

	Register("GET", "/tail/*", func(c *Context) {
		c.SetContentType("text/plain; charset=utf-8")
		c.WriteHeader(200)
		c.WriteString(c.Param("*"))
	})

	RegisterAuth("GET", "/auth", func(c *Context) {
		c.SetContentType("text/plain; charset=utf-8")
		c.WriteHeader(200)
		c.WriteString("OK")
	})

	RegisterJsonRPC("/rpc")

	RegMethod("hello", func(name string) (string, *jrpc.Error) {
		return "Hello, " + name + "!", nil
	})

	File("/LICENSE", "./LICENSE")

	tG := func(url string, code int, body string) {

		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		rt.ServeHTTP(rw, req)

		res := rw.Result()

		if res == nil || res.StatusCode != code {
			t.Fatalf("Invalid reponse for: %s", url)
		}

		if res.StatusCode == 200 {

			data, _ := ioutil.ReadAll(res.Body)

			if string(data) != body {
				t.Fatalf("Invalid body for: %s", url)
			}
		}
	}

	tJRPC := func(method string, param interface{}, ret string) {

		params := make(map[string]interface{})

		params["jsonrpc"] = "2.0"
		params["method"] = method
		params["id"] = 123
		params["params"] = param

		data, _ := json.Marshal(params)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/rpc", bytes.NewReader(data))

		rt.ServeHTTP(rw, req)

		res := rw.Result()

		if res == nil || res.StatusCode != 200 {
			t.Fatal("Invalid reponse for: /rpc")
		}

		if res.StatusCode == 200 {

			data, _ := ioutil.ReadAll(res.Body)
			if strings.Index(string(data), ret) < 0 {
				t.Fatal("Template not found")
			}
		}
	}

	SetUID(true)

	tG("/", 200, "Hello!")
	tG("/unknown", 404, "404 unknown request")
	tG("//", 200, "Hello!")
	tG("/user/wmentor", 200, "Hello, wmentor!")
	tG("/user/wmentor/", 200, "Hello, wmentor!")
	tG("/tail/", 404, "404 unknown request")
	tG("/tail/1", 200, "/1")
	tG("/tail/1/11/111/", 200, "/1/11/111")
	tG("/auth", 401, "Auth require.")
	tG("/LICENSE", 200, `MIT License

Copyright (c) 2020 Mikhail Kirillov <wmentor@mail.ru>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`)

	tJRPC("hello", "wmentor", `"result":"Hello, wmentor!"`)
}
