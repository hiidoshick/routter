package routter

import (
	"log"
	"net/http"
	"strings"
	"time"
)

var Session *http.Server

type Keys map[string]string
type Router struct {
	mux      map[string]Handle
	urls     []string
	NotFound func(http.ResponseWriter, *http.Request)
	Host     string
}
type Handle func(http.ResponseWriter, *http.Request, Keys)

func (r *Router) Run(port string) {
	Session = &http.Server{
		Addr:           port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Running server...")
	log.Fatal(Session.ListenAndServe())
}

func New() *Router {
	return &Router{
		mux:      make(map[string]Handle),
		urls:     make([]string, 0),
		NotFound: http.NotFound,
		Host:     "localhost",
	}
}

func (r *Router) Add(path string, handle Handle) {
	log.Println("PATH: ", path)
	r.mux[path] = handle
	r.urls = append(r.urls, path)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fullURL := req.Host + req.URL.Path
	head, parsed := r.satisfyRegexp(fullURL)
	log.Println("Routter: ", head, parsed, fullURL)
	h, ok := r.mux[head]
	if ok && parsed != nil {

		h(w, req, parsed)
		return
	}
	r.NotFound(w, req)
}

func (r *Router) parseRegExp(regexp string, str string) Keys {
	re := strings.Split(regexp, "/")
	st := strings.Split(str, "/")
	re[0] = strings.Split(regexp, "."+r.Host)[0]
	if strings.Contains(re[0], r.Host) {
		re[0] = ""
	}
	st[0] = strings.Split(str, "."+r.Host)[0]
	if strings.Contains(st[0], r.Host) {
		st[0] = ""
	}
	if strings.TrimRight(regexp, ":") == regexp && len(re) != len(st) {
		log.Println("NIL:", re, st)
		return nil
	}
	var p Keys
	p = make(Keys)
	for i := range re {
		if !strings.Contains(re[i], "$") && re[i] != st[i] {
			return nil
		}
		if strings.Trim(re[i], " \n") != "" && strings.Trim(st[i], " \n") == "" {
			return nil
		}
		if strings.Contains(re[i], "$") {
			p[strings.Replace(re[i], "$", "", 1)] = st[i]
			if strings.Contains(re[i], ":") {
				p[strings.Replace(strings.Replace(re[i], "$", "", 1), ":", "", 1)] = str[strings.Index(str, st[i]):]
				delete(p, strings.Replace(re[i], "$", "", 1))
			}
		} else {
			p[re[i]] = st[i]
		}
	}
	return p

}

func (r *Router) satisfyRegexp(url string) (string, Keys) {
	for _, e := range r.urls {
		if parsed := r.parseRegExp(e, url); parsed != nil {
			return e, parsed
		}
	}
	return "", nil
}
