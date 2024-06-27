package proxy

import (
	"bytes"
	"crypto/tls"
	logger "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"security-gateway/pkg/util"
)

type ModifiableResponseWriter struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (m *ModifiableResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *manager) getProxyService(targetUrl string) (*httputil.ReverseProxy, error) {
	rp, ok := m.proxyServices[targetUrl]
	if !ok {
		tu, err := url.Parse(targetUrl)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		rp = httputil.NewSingleHostReverseProxy(tu)

		rp.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		rp.Director = func(req *http.Request) {
			targetQuery := tu.RawQuery
			req.URL.Scheme = tu.Scheme
			req.URL.Host = tu.Host
			req.Host = tu.Host
			req.URL.Path, req.URL.RawPath = util.JoinURLPath(tu, req.URL)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
		}

		m.proxyServices[targetUrl] = rp
	}
	return rp, nil
}
