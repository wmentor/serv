package serv

import (
	"net/http/httptest"
	"testing"
)

func TestQuery(t *testing.T) {

	req := httptest.NewRequest("GET", "/123/13241234?test=1&login=monster&pt=true&pt=p2&wsdl&f=false", nil)
	if req == nil {
		t.Fatal("test request create failed")
	}

	qw := newQuery(req)
	if qw == nil {
		t.Fatal("newQuery failed")
	}

	if qw.Has("unknown") {
		t.Fatal("Has return true for unknown param")
	}

	if !qw.Has("test") || !qw.Has("login") || !qw.Has("wsdl") {
		t.Fatal("Not found all fields")
	}

	if qw.GetInt64("unknown") != 0 || qw.GetInt64("test") != 1 || qw.GetInt64("wsdl") != 0 {
		t.Fatal("GetInt64 failed")
	}

	if qw.GetInt("unknown") != 0 || qw.GetInt("test") != 1 || qw.GetInt("wsdl") != 0 {
		t.Fatal("GetInt failed")
	}

	if qw.GetFloat("unknown") != 0 || qw.GetFloat("test") != 1 || qw.GetFloat("wsdl") != 0 {
		t.Fatal("GetFloat failed")
	}

	if qw.GetString("unknown") != "" || qw.GetString("test") != "1" || qw.GetString("wsdl") != "" || qw.GetString("f") != "false" {
		t.Fatal("GetFloat failed")
	}

	if qw.GetBool("unknown") || !qw.GetBool("test") || qw.GetBool("wsdl") || qw.GetBool("f") || !qw.GetBool("test") || !qw.GetBool("pt") {
		t.Fatal("GetBool failed")
	}
}
