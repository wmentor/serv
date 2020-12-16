package serv

import (
	"net/http"
	"strconv"
)

type Query map[string][]string

func newQuery(req *http.Request) Query {
	param := req.URL.Query()
	return Query(param)
}

func (p Query) Has(name string) bool {
	_, has := p[name]
	return has
}

func (p Query) GetInt64(name string) int64 {
	data, has := p[name]

	if !has || data == nil || len(data) == 0 {
		return 0
	}

	v, err := strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return 0
	}

	return v
}

func (p Query) GetInt(name string) int {
	data, has := p[name]

	if !has || data == nil || len(data) == 0 {
		return 0
	}

	v, err := strconv.Atoi(data[0])
	if err != nil {
		return 0
	}

	return v
}

func (p Query) GetBool(name string) bool {

	data, has := p[name]
	if !has || data == nil || len(data) == 0 {
		return false
	}

	v, err := strconv.ParseBool(data[0])
	if err != nil {
		v = false
	}

	return v
}

func (p Query) GetString(name string) string {
	data, has := p[name]
	if !has || data == nil || len(data) == 0 {
		return ""
	}

	return data[0]
}

func (p Query) GetFloat(name string) float64 {
	data, has := p[name]
	if !has || data == nil || len(data) == 0 {
		return 0
	}

	v, err := strconv.ParseFloat(data[0], 64)
	if err != nil {
		return 0
	}

	return v
}
