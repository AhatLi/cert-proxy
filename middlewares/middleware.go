package middlewares

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy-gateway/config"
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

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func GinBodyLogMiddleware(c *gin.Context) {
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	if c.Request.Body != nil {
		b, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			// handle error
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}

	fmt.Println("c.Writer", blw.body.String())
	fmt.Println("c.Writer.Status()", c.Writer.Status())
}
