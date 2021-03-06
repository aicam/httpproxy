package main

import (
	"encoding/json"
	"flag"
	"github.com/aicam/httpproxy/server"
	"github.com/aicam/jsonconfig"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var Config jsonconfig.Configuration

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
}
type LogFormat struct {
	HOST       string `json:"host"`
	Path       string `json:"path"`
	Fragment   string `json:"fragment"`
	CategoryID uint   `json:"category_id"`
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	//log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)
	var category uint
	for i, element := range Categories {
		if strings.Contains(req.URL.Host, element) {
			category = i
		}
	}
	js, _ := json.Marshal(LogFormat{
		HOST:       req.URL.Host,
		Path:       req.URL.Path,
		Fragment:   req.URL.Fragment,
		CategoryID: category,
	})
	_, _ = File.Write(append(js, byte(',')))
	_, _ = File.Write([]byte("\n"))
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		msg := "unsupported protocal scheme " + req.URL.Scheme
		http.Error(wr, msg, http.StatusBadRequest)
		log.Println(msg)
		return
	}

	client := &http.Client{}

	req.RequestURI = ""

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Print("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	//log.Println(req.RemoteAddr, " ", resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

var File *os.File
var Categories map[uint]string

func main() {
	var addr = flag.String("addr", "127.0.0.1:8080", "The addr of the application.")
	flag.Parse()
	handler := &proxy{}
	Categories = make(map[uint]string)
	Config = jsonconfig.Read("./user-config.json")
	for _, conf := range Config.SitesCategories {
		Categories[conf.CategoryId] = conf.HOST
	}
	File, _ = os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.Println("Starting proxy server on", *addr)
	go http.ListenAndServe(*addr, handler)
	router := gin.Default()
	router.Use(cors.Default())
	var configJS jsonconfig.Configuration
	router.GET("/read_config-file", func(context *gin.Context) {
		context.String(http.StatusOK, server.ReadConfig("./user-config.json"))
	})
	router.POST("/write_config-file", func(context *gin.Context) {
		context.BindJSON(&configJS)
		server.WriteConfig("./user-config.json", configJS)
		context.JSON(http.StatusOK, struct {
			Status bool `json:"status"`
		}{Status: true})
	})
	router.GET("/info", func(context *gin.Context) {
		context.String(http.StatusOK, string(server.GetInfo(Categories, Config.Config.FileName, Config)))
	})
	router.NoRoute(func(context *gin.Context) {
		dir, file := path.Split(context.Request.RequestURI)
		ext := filepath.Ext(file)
		if file == "" || ext == "" {
			context.File("./dist/httpproxy-front/index.html")
		} else {
			context.File("./dist/httpproxy-front/" + path.Join(dir, file))
		}
	})
	_ = router.Run("0.0.0.0:4300")
}
