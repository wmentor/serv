package serv

import (
	"time"
)

type LogData struct {
	Method     string
	Addr       string
	Auth       string
	RequestURL string
	StatusCode int
	Seconds    float64
	Referer    string
	UserAgent  string
	UID        string
}

type Logger func(*LogData)

type LongQueryHandler func(delta time.Duration, c *Context)
