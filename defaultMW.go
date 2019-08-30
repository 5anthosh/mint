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
	rootResponse := make(map[string]interface{})
	errResponse := make(map[string]interface{})
	errResponse["code"] = http.StatusNotFound
	errResponse["message"] = "Resource not found"
	rootResponse["error"] = errResponse
	c.JSON(http.StatusNotFound, rootResponse)
}

func methodNotAllowedHandler(c *Context) {
	rootResponse := make(map[string]interface{})
	errResponse := make(map[string]interface{})
	errResponse["code"] = http.StatusMethodNotAllowed
	errResponse["message"] = "Method not allowed"
	rootResponse["error"] = errResponse
	c.JSON(http.StatusMethodNotAllowed, rootResponse)
}
