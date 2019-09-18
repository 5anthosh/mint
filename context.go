package mint

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

//constant
const (
	emptyString = ""
	PayloadKey  = "PayLoad"
)

var (
	newLine         = []byte{'\n'}
	jsonContentType = []string{"application/json; charset=utf-8"}
)

//Context provides context for whole request/response cycle
//It helps to pass variable from one middlware to another
type Context struct {
	*HandlerContext
	Request   *http.Request
	Response  http.ResponseWriter
	Method    string
	URLParams map[string]string
	index     int
	status    int
	size      int
	errors    []error
}

func (app *Mint) newContext() *Context {
	return new(Context)
}

//Reset resets the value the context
func (c *Context) Reset() {
	c.HandlerContext = nil
	c.Request = nil
	c.Response = nil
	c.status = 0
	c.size = 0
	c.errors = c.errors[0:0]
	c.index = 0
}
func newContextPool(app *Mint) func() interface{} {
	return func() interface{} {
		return app.newContext()
	}
}

//GetHeader returns request header
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

//Status set http status code
func (c *Context) Status(status int) {
	c.status = status
	c.Response.WriteHeader(status)
}

func (c *Context) writeContentType(contentType []string) {
	c.SetHeader("Content-Type", contentType)
}

//SetHeader #
func (c *Context) SetHeader(key string, value []string) {
	header := c.Response.Header()
	if val := header[key]; len(val) == 0 {
		header[key] = value
	}
}

//AddHeader add header to the response
func (c *Context) AddHeader(key string, value string) {
	c.Response.Header().Add(key, value)
}
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

//CJSON writes compressed json response
func (c *Context) compressedJSON(code int, reponse interface{}) {
	// create header
	c.SetHeader("Content-Encoding", []string{"gzip"})
	// Gzip data
	c.Status(code)
	gz := c.Mint.gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(c.Response)
	jsonContentByte, err := json.Marshal(reponse)
	if err != nil {
		c.Errors(err)
	}

	size, err := gz.Write(jsonContentByte)
	if err != nil {
		c.Errors(err)
	}
	c.setSize(size)
	size, err = gz.Write(newLine)
	c.setSize(size)

	if err != nil {
		c.Errors(err)
	}
	gz.Close()
	c.Mint.gzipWriterPool.Put(gz)
}

//JSON #
func (c *Context) JSON(code int, response interface{}) {
	c.SetHeader("Content-Type", jsonContentType)
	if bodyAllowedForStatus(code) {
		if c.HandlerContext.compressed {
			c.compressedJSON(code, response)
		} else {
			c.uncompressedJSON(code, response)
		}
	} else {
		c.Status(code)
	}
}

//JSON writes json response
func (c *Context) uncompressedJSON(code int, reponse interface{}) {
	c.Status(code)
	jsonContentByte, err := json.Marshal(reponse)
	if err != nil {
		c.Errors(err)
	}
	size, err := c.Response.Write(jsonContentByte)
	if err != nil {
		c.Errors(err)
	}
	c.setSize(size)
	size, err = c.Response.Write(newLine)
	c.setSize(size)
	if err != nil {
		c.Errors(err)
	}

}

func (c *Context) setSize(size int) {
	c.size += size
}

//Errors records error to be displayed later
func (c *Context) Errors(err ...error) {
	for _, er := range err {
		c.Error(er)
	}
}

func (c *Context) Error(err error) {
	if err != nil {
		c.errors = append(c.errors, err)
	}
}

//ClientIP returns ip address of the user using request info
func (c *Context) ClientIP() string {
	clientIP := c.GetHeader("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == emptyString {
		clientIP = strings.TrimSpace(c.GetHeader("X-Real-Ip"))
	}
	if clientIP != emptyString {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return emptyString
}

//Next runs the next handler
func (c *Context) Next() {
	if c.index >= c.HandlerContext.count {
		return
	}
	handle := c.HandlerContext.handlers[c.index]
	c.index++
	handle(c)
}

//GetURLQuery get the params in url (Eg . /?q=)
func (c *Context) GetURLQuery(query string) string {
	return c.Request.URL.Query().Get(query)
}

//Get #
func (c *Context) Get(key interface{}) interface{} {
	return c.Request.Context().Value(key)
}

//GetString gets string value using key in context
func (c *Context) GetString(key interface{}) string {
	return c.Get(key).(string)
}

//GetInt64 gets values associated with key as int64
func (c *Context) GetInt64(key interface{}) int64 {
	return c.Get(key).(int64)
}

//GetFloat64 gets values associated with key as float64
func (c *Context) GetFloat64(key interface{}) float64 {
	return c.Get(key).(float64)
}

//GetComplex128 gets values associated with key as complex128
func (c *Context) GetComplex128(key interface{}) complex128 {
	return c.Get(key).(complex128)
}

//Set #
func (c *Context) Set(key, val interface{}) {
	if val == nil {
		return
	}
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), key, val))
}

// URI gets request uri
func (c *Context) URI() string {
	return c.Request.RequestURI
}
