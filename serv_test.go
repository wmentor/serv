package serv

import (
	"strings"
	"testing"
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
