package serv

import (
	"strconv"
)

type Params map[string]string

func (p Params) GetString(name string) string {
	if v, h := p[name]; h {
		return v
	}

	return ""
}

func (p Params) GetInt(name string) int {
	if v, h := p[name]; h {
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}

		return 0
	}

	return 0
}

func (p Params) GetInt64(name string) int64 {
	if v, h := p[name]; h {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			return val
		}

		return 0
	}

	return 0
}

func (p Params) GetFloat(name string) float64 {
	if v, h := p[name]; h {
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			return val
		}

		return 0
	}

	return 0
}

func (p Params) GetBool(name string) bool {
	if v, h := p[name]; h {
		if val, err := strconv.ParseBool(v); err == nil {
			return val
		}

		return false
	}

	return false
}
