package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"razer-ford/proxy-server/internal/pkg/cache"
	"strconv"
	"time"
)

var (
	glLog = log.Default()

	errEtagNotFound = errors.New("Etag not found")
)

const (
	green = "\033[0;32m"
	reset = "\033[0m"

	Etag         = "Etag"
	CacheControl = "Cache-Control"
	LastModified = "Last-Modified"

	IfModifiedSince = "If-Modified-Since"
	IfNoneMatch     = "If-None-Match"

	layout = "Mon, 1 Jan 2006 00:00:00 GMT"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func strToGreen(str string) string {
	return fmt.Sprintf("%v%v%v", green, str, reset)
}

type ProxyServer struct {
	port    int
	address string
	journal log.Logger
	ch      *cache.Cache
}

func NewProxyServer(port int, address string) *ProxyServer {
	return &ProxyServer{
		port:    port,
		address: address,
		journal: *log.New(os.Stdout, "", log.Ldate|log.Ltime),
		ch:      cache.NewCache(),
	}
}

func (ps *ProxyServer) Run() error {
	glLog.Printf("running on address http://%v:%v\n", ps.address, ps.port)

	http.HandleFunc("/favicon.ico", faviconHandler)

	return http.ListenAndServe(ps.address+":"+strconv.Itoa(ps.port), NewProxyServer(ps.port, ps.address))
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		{
			p.handlePost(w, req)
		}
	case http.MethodGet:
		{
			p.handleGet(w, req)
		}
	default:
		{
			glLog.Printf("was expected %v or %v but was %v", http.MethodGet, http.MethodPost, req.Method)
		}
	}
}

func (p *ProxyServer) handlePost(w http.ResponseWriter, req *http.Request) {
	p.handle(w, req)
}

func (p *ProxyServer) handleGet(w http.ResponseWriter, req *http.Request) {
	p.handle(w, req)
}

func (p *ProxyServer) handle(w http.ResponseWriter, req *http.Request) {
	client := http.Client{}

	newReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	newReq.Header = req.Header.Clone()

	if err != nil {
		glLog.Printf("error creating Request for proxy: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	resp, err := client.Do(newReq)
	if err != nil {
		glLog.Printf("the request failed: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		glLog.Printf("body reading error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = p.cached(resp, body[:]); err != nil {
		glLog.Println(err)
	}

	copyHeader(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	p.journal.Printf(strToGreen("{URL: %v; Status: %v}"), resp.Request.URL, resp.Status)
}

func (p *ProxyServer) cached(resp *http.Response, body []byte) error {
	h := &resp.Header

	etag := h.Get(Etag)
	if etag == "" {
		return errEtagNotFound
	}
	t, err := time.Parse(layout, h.Get(LastModified))
	glLog.Println(h.Get(LastModified))
	if err != nil {
		return err
	}
	data := cache.NewData(
		etag[1:len(etag)-1],
		&t,
		h.Get(CacheControl),
		body,
	)
	p.ch.Set(data.Key, data)
	return nil
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "assets/favicon.ico")
}

func copyHeader(src, dst http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
