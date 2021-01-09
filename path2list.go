package serv

import (
	"net/url"
	"strings"
)

func path2list(path string) []string {

	if len(path) < 1 || path[0] != '/' {
		return nil
	}

	list := make([]string, 1, 32)

	list[0] = "/"

	for _, v := range strings.Split(path[1:], "/") {
		if v != "" {
			if uri, e := url.PathUnescape(v); e == nil {
				list = append(list, uri)
				if uri == "*" {
					return list
				}
			} else {
				return nil
			}

		}
	}

	return list
}
