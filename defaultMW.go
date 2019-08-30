package mint

import (
	"net/http"
	"time"
)

//LoggerMW logger middleware
func loggerMW(c *Context) {
	start := time.Now()
	path := c.Request.URL.Path
	c.Next()
	log := new(Logger)
	log.TimeStamp = time.Now()
	log.Latency = log.TimeStamp.Sub(start)
	log.Method = c.Method
	log.StatusCode = c.status
	log.ClientIP = c.ClientIP()
	log.BodySize = c.size
	log.Path = path
	log.Errors = c.errors
	log.Print()
}

func notFoundHandler(c *Context) {
	c.Status(http.StatusNotFound)
}

func methodNotAllowedHandler(c *Context) {
	c.Status(http.StatusMethodNotAllowed)
}
