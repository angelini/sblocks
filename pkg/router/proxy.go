package router

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/angelini/sblocks/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type Proxy struct {
	port       int
	httpClient *http.Client
}

func NewProxy(port int) (*Proxy, error) {
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	return &Proxy{
		port: port,
		httpClient: &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}, nil
}

func (p *Proxy) Start(ctx context.Context) error {
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		// TODO
		hostname := ""

		body, err := io.ReadAll(req.Body)
		if err != nil {
			httpErr(ctx, resp, err, "failed to read proxy request body")
			return
		}

		url := fmt.Sprintf("http://%s%s", hostname, req.URL.String())
		proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))
		if err != nil {
			httpErr(ctx, resp, err, "failed to create proxy request")
			return
		}

		proxyReq.Header = make(http.Header)
		copyHeader(proxyReq.Header, req.Header, true)

		remoteHost, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			appendHostToXForwardHeader(req.Header, remoteHost)
		}

		proxyResp, err := p.httpClient.Do(proxyReq)
		if err != nil {
			httpErr(ctx, resp, err, "failed to proxy request")
			return
		}
		defer proxyResp.Body.Close()

		copyHeader(resp.Header(), proxyResp.Header, false)
		resp.WriteHeader(proxyResp.StatusCode)
		io.Copy(resp, proxyResp.Body)
	})

	return http.ListenAndServe(":"+strconv.Itoa(p.port), nil)
}

func httpErr(ctx context.Context, resp http.ResponseWriter, err error, message string) {
	log.Error(ctx, message, zap.Error(err))
	http.Error(resp, err.Error(), http.StatusInternalServerError)
}

// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true, // canonicalized version of "TE"
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

func copyHeader(dest, src http.Header, skipHopHeaders bool) {
	for key, value := range src {
		if skipHopHeaders {
			if _, ok := hopHeaders[key]; ok {
				continue
			}
		}

		for _, nested := range value {
			dest.Add(key, nested)
		}
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
