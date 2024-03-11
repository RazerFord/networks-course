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

	http.HandleFunc("/", NewProxyServer(ps.port, ps.address).httpHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	return http.ListenAndServe(ps.address+":"+strconv.Itoa(ps.port), nil)
}

func (p *ProxyServer) httpHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		glLog.Printf("was expected %v but was %v", http.MethodGet, req.Method)
		return
	}
	// glLog.Println("start")
	// glLog.Printf("scheme %v\n", req.URL.Scheme)
	// glLog.Printf("port %v\n", req.URL.Port())
	// glLog.Printf("uri: %v\n", req.RequestURI)
	// glLog.Printf("addr: %v\n", req.RemoteAddr)
	// glLog.Printf("header: %v\n", req.Header)
	// glLog.Printf("host: %v\n", req.Host)
	// glLog.Printf("url-host: %v\n", req.URL.Host)

	client := http.Client{}
	nreq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	req.Body.Close()

	if err != nil {
		glLog.Printf("error creating Request for proxy: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nreq.RequestURI = nreq.RequestURI
	nreq.Header = req.Header.Clone()

	resp, err := client.Do(nreq)
	if err != nil {
		glLog.Printf("the request failed: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	// glLog.Println("start")
	// glLog.Printf("scheme %v\n", nreq.URL.Scheme)
	// glLog.Printf("port %v\n", nreq.URL.Port())
	// glLog.Printf("uri: %v\n", nreq.RequestURI)
	// glLog.Printf("addr: %v\n", nreq.RemoteAddr)
	// glLog.Printf("header: %v\n", nreq.Header)
	// glLog.Printf("host: %v\n", nreq.Host)
	// glLog.Printf("url-host: %v\n", nreq.URL.Host)

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
