package hime

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// Param is the query param when redirect
type Param struct {
	Name  string
	Value any
}

// SafeRedirectPath filters domain out from path
func SafeRedirectPath(p string) string {
	l, err := url.ParseRequestURI(p)
	if err != nil {
		return "/"
	}
	r := l.EscapedPath()
	if len(r) == 0 {
		r = "/"
	}
	if l.ForceQuery || l.RawQuery != "" {
		r += "?" + l.RawQuery
	}
	return path.Clean(r)
}

func mergeValues(s, p url.Values) {
	for k, v := range p {
		for _, vv := range v {
			s[k] = append(s[k], vv)
		}
	}
}

func mergeValueWithMapString(s url.Values, m map[string]string) {
	for k, v := range m {
		s[k] = append(s[k], v)
	}
}

func mergeValueWithMapAny(s url.Values, m map[string]any) {
	for k, v := range m {
		s[k] = append(s[k], fmt.Sprint(v))
	}
}

func buildPath(base string, params ...any) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		panicf("parse url error; %v", err)
	}

	xs := make([]string, 0, len(params))
	ps := make(url.Values)
	mergeValues(ps, baseURL.Query())
	for _, p := range params {
		switch v := p.(type) {
		case url.Values:
			mergeValues(ps, v)
		case map[string]string:
			mergeValueWithMapString(ps, v)
		case map[string]any:
			mergeValueWithMapAny(ps, v)
		case *Param:
			ps[v.Name] = append(ps[v.Name], fmt.Sprint(v.Value))
		default:
			xs = append(xs, strings.TrimPrefix(fmt.Sprint(p), "/"))
		}
	}
	if baseURL.Path == "" || (len(xs) > 0 && !strings.HasSuffix(baseURL.Path, "/")) {
		baseURL.Path += "/"
	}
	baseURL.Path += path.Join(xs...)
	baseURL.RawQuery = ps.Encode()
	return baseURL.String()
}
