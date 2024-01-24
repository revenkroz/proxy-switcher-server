package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	// Server configuration
	listenAddr = getFromEnvString("LISTEN_ADDR", "0.0.0.0:8888")

	// Proxy configuration
	targetUrlRaw            = getFromEnvString("TARGET_URL", "")
	proxyList    arrayFlags = getFromEnvStringSlice("PROXY_LIST")
	// Status codes to trigger proxy switch
	triggerCodes arrayFlags = getFromEnvStringSlice("TRIGGER_CODES")

	// Internal variables
	targetUrl         *url.URL
	currentProxyIndex = 0
	currentProxy      *url.URL
	proxyMutex        sync.Mutex
)

func init() {
	flag.StringVar(&listenAddr, "listen", listenAddr, "Server listen address (if empty, env:LISTEN_ADDR will be used).")
	flag.StringVar(&targetUrlRaw, "target", targetUrlRaw, "Target URL (if empty, env:TARGET_URL will be used).")
	flag.Var(&proxyList, "proxy", "List of proxies (if empty, env:PROXY_LIST will be used).")
	flag.Var(&triggerCodes, "trigger-code", "List of status codes to trigger proxy switch (if empty, env:TRIGGER_CODES will be used).")
	flag.Parse()

	if len(proxyList) == 0 {
		log.Fatalln("No proxies provided")
	}

	if len(triggerCodes) == 0 {
		triggerCodes = append(triggerCodes, "429")
	}

	targetUrl = parseUrl(targetUrlRaw)
	currentProxy = parseUrl(proxyList[currentProxyIndex])
}

func main() {
	http.HandleFunc("/", handleRequest)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting proxy server on %s\n", listenAddr)
	log.Printf("Proxying to %s\n", targetUrl)
	log.Printf("> Using %s as the first proxy\n", currentProxy)

	server := &http.Server{}
	err = server.Serve(ln)
	if err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ww := newWriterWrapper(w)

	proxy := httputil.NewSingleHostReverseProxy(currentProxy)
	proxy.Transport = createTransport()

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = targetUrl.Host
	}

	retriesCount := 0

	proxy.ServeHTTP(ww, r)

	// Handle rate limit
	for checkTriggerCode(ww.Status()) {
		log.Printf(">> Rate limited through %s\n", currentProxy)

		retriesCount++
		if retriesCount >= len(proxyList) {
			log.Println("[!] All proxies are rate limited.")
			break
		}

		newProxy := updateCurrentProxy()
		log.Printf("> Switching to the next proxy: %s\n", newProxy)

		proxy.Transport = createTransport()

		ww.Reset()
		proxy.ServeHTTP(ww, r)
	}

	// it could be caused by the proxy server
	if ww.Status() == http.StatusBadGateway {
		log.Println("[!] Bad gateway. It could be caused by the proxy server.")
	}

	ww.SendResponse()
}

func updateCurrentProxy() string {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	currentProxyIndex = (currentProxyIndex + 1) % len(proxyList)
	currentProxy = parseUrl(proxyList[currentProxyIndex])

	return proxyList[currentProxyIndex]
}

func createTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyURL(currentProxy),
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
	}
}

func checkTriggerCode(code int) bool {
	codeStr := strconv.Itoa(code)

	for _, triggerCode := range triggerCodes {
		if triggerCode == "*" {
			return true
		}

		if triggerCode == codeStr {
			return true
		}
	}

	return false
}
