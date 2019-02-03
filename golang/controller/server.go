package controller

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Frizz925/gbf-proxy/golang/lib"
)

type Server struct {
	base *lib.BaseServer
}

func NewServer() lib.Server {
	return &Server{
		base: lib.NewBaseServer("Controller"),
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
	return s.base.Open(addr, s.serve)
}

func (s *Server) Close() (bool, error) {
	return s.base.Close()
}

func (s *Server) WaitGroup() *sync.WaitGroup {
	return s.base.WaitGroup
}

func (s *Server) Listener() net.Listener {
	return s.base.Listener
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	host := req.Host
	hostname := host
	tokens := strings.SplitN(host, ":", 2)
	if len(tokens) >= 2 {
		hostname = tokens[0]
	}
	if !strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		writeError(w, 403, "Host not allowed")
		return
	}

	url, err := url.Parse(req.RequestURI)
	if err != nil {
		writeError(w, 400, "Bad request URI")
		return
	}
	if url.Scheme == "" {
		url.Scheme = "http"
	}
	url.Host = host

	c := http.Client{}
	res, err := c.Do(&http.Request{
		Method: req.Method,
		URL:    url,
		Body:   req.Body,
		Header: req.Header,
	})
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	for k, values := range res.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.StatusCode)
	length := len(body)
	for written := 0; written < length; {
		write, err := w.Write(body[written:])
		if err != nil {
			panic(err)
		}
		written += write
	}
}

func (s *Server) serve(l net.Listener) {
	err := http.Serve(l, s)
	if err != nil {
		// do nothing
	}
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(message + "\r\n"))
	if err != nil {
		panic(err)
	}
}