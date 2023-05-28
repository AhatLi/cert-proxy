package middlewares

import (
	"bytes"
	"cert-proxy/config"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func AccessControlAllowOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", config.GetConfig().Access_control_allow_origin)
		c.Header("Access-Control-Allow-Headers", config.GetConfig().Access_control_allow_headers)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	}
}

func ReturnReverseProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		conf := config.GetConfig()
		_, port, _ := net.SplitHostPort(c.Request.Host)
		confUrl := conf.Network_list[port]
		if confUrl == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		fmt.Println("port = "+port, " URL = "+confUrl)

		target, err := url.Parse(confUrl)
		if err != nil {
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ModifyResponse = func(res *http.Response) error {
			location := res.Header.Get("Location")
			if strings.Contains(location, "http:") {
				res.Header.Set("Location", strings.Replace(location, "http:", "https:", 1))
			}
			return nil
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

type BodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r bodyLogWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func GinBodyLogMiddleware(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()

	body := blw.body.String()
	newBody := strings.ReplaceAll(body, "http://"+c.Request.Host, "https://"+c.Request.Host)

	blw.body = &bytes.Buffer{}
	blw.Write([]byte(newBody))
	blw.ResponseWriter.Write(blw.body.Bytes())
}
