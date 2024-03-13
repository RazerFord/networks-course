package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	glLog = log.Default()
)

const (
	green = "\033[0;32m"
	reset = "\033[0m"
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
}

func NewProxyServer(port int, address string) *ProxyServer {
	return &ProxyServer{
		port:    port,
		address: address,
		journal: *log.New(os.Stdout, "", log.Ldate|log.Ltime),
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

		}
	case http.MethodGet:
		{
			p.handleGet(w, req)
		}
	default:
		{
			glLog.Printf("was expected %v but was %v", http.MethodGet, req.Method)
		}
	}
}

func (p *ProxyServer) handleGet(w http.ResponseWriter, req *http.Request) {
	client := http.Client{}

	newReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	req.Body.Close()
	newReq.Header = req.Header.Clone()

	if err != nil {
		glLog.Printf("error creating Request for proxy: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

	header := w.Header()
	copyHeader(resp.Header, header)
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	p.journal.Printf(strToGreen("{URL: %v; Status: %v}"), resp.Request.URL, resp.Status)
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
