package proxy

import (
	"crypto/tls"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/valyala/fasthttp"
)

func init() {
	proxy.WithClient(&fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
}
