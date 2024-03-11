package server

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

var glLog = log.Default()

type ProxyServer struct {
	port    int
	address string
	journal log.Logger
}

func NewProxyServer(port int, address string) *ProxyServer {
	return &ProxyServer{
		port:    port,
		address: address,
		journal: *log.Default(),
	}
}

func (ps *ProxyServer) Run() error {
	glLog.Printf("running on address http://%v:%v\n", ps.address, ps.port)

	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	return http.ListenAndServe(ps.address+":"+strconv.Itoa(ps.port), nil)
}

func httpHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		if req.Method != http.MethodGet {
			glLog.Printf("was expected %v but was %v", http.MethodGet, req.Method)
			return
		}

		client := http.Client{}

		rawUrl := req.RequestURI[1:]
		if req.URL.Scheme == "" {
			if req.URL.Port() == "443" {
				rawUrl = "https://" + rawUrl
			} else {
				rawUrl = "http://" + rawUrl
			}
		} else {
			rawUrl = req.URL.Scheme + rawUrl
		}

		http.Get(req.RequestURI)
		nreq, err := http.NewRequest(req.Method, rawUrl, req.Body)

		if err != nil {
			glLog.Println(err)
			return
		}

		resp, err := client.Do(nreq)

		if err != nil {
			glLog.Println(err)
			return
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		w.Write(body)
		w.WriteHeader(resp.StatusCode)

		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "assets/favicon.ico")
}
