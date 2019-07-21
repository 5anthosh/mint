package mint

import (
	"compress/gzip"
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

var (
	newLine     = []byte{'\n'}
	emptyString = ""
)
var (
	jsonContentType  = []string{"application/json; charset=utf-8"}
	cjsonContentType = []string{"application/json"}
)

//Context provides context for whole request/response cycle
type Context struct {
	*HandlersContext
	Request    *http.Request
	Response   http.ResponseWriter
	store      map[string]interface{}
	URLParams  map[string]string
	Params     map[string]string
	index      int8
	StatusCode int
	Size       int
	Error      []error
	CR         bool
}

func (app *Mint) newContext() *Context {
	return &Context{
		Params: make(map[string]string),
	}
}

//Reset resets the value the context
func (c *Context) Reset() {
	c.HandlersContext = nil
	c.Request = nil
	c.Response = nil
	c.Params = make(map[string]string)
	c.store = make(map[string]interface{})
	c.StatusCode = 0
	c.Size = 0
	c.Error = c.Error[0:0]
	c.index = 0
	c.CR = false
}
func newContextPool(app *Mint) func() interface{} {
	return func() interface{} {
		return app.newContext()
	}
}

//Get #
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

//Set #
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

//GetRequestHeader returns request header
func (c *Context) GetRequestHeader(key string) string {
	return c.Request.Header.Get(key)
}

//HTTPStatus set http status code
func (c *Context) HTTPStatus(status int) {
	c.StatusCode = status
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

//AddHeader #
func (c *Context) AddHeader(key string, value string) {
	c.Response.Header().Set(key, value)
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
func (c *Context) compressedJSON(reponse interface{}) {
	// create header
	c.writeContentType(jsonContentType)
	c.AddHeader("Content-Encoding", "gzip")
	// Gzip data
	gz := c.mint.gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(c.Response)
	jsonContentByte, err := json.Marshal(reponse)
	if err != nil {
		c.AppendError(err)
	}
	size, err := gz.Write(jsonContentByte)
	if err != nil {
		c.AppendError(err)
	}
	c.setSize(size)
	size, err = gz.Write(newLine)
	c.setSize(size)
	if err != nil {
		c.AppendError(err)
	}
	gz.Close()
	c.mint.gzipWriterPool.Put(gz)
}

//JSON #
func (c *Context) JSON(code int, response interface{}) {
	c.HTTPStatus(code)
	if bodyAllowedForStatus(code) {
		if c.CR {
			c.compressedJSON(response)
		} else {
			c.uncompressedJSON(response)
		}
	}
}

//JSON writes json response
func (c *Context) uncompressedJSON(reponse interface{}) {
	c.Response.Header().Add("Content-Type", "application/json")
	jsonContentByte, err := json.Marshal(reponse)
	if err != nil {
		c.AppendError(err)
	}
	size, err := c.Response.Write(jsonContentByte)
	if err != nil {
		c.AppendError(err)
	}
	c.setSize(size)
	size, err = c.Response.Write(newLine)
	c.setSize(size)
	if err != nil {
		c.AppendError(err)
	}

}

func (c *Context) setSize(size int) {
	c.Size += size
}

//AppendError records error to be displayed later
func (c *Context) AppendError(err ...error) {
	if err != nil {
		c.Error = append(c.Error, err...)
	}
}

//ClientIP returns ip address of the user using request info
func (c *Context) ClientIP() string {
	clientIP := c.GetRequestHeader("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == emptyString {
		clientIP = strings.TrimSpace(c.GetRequestHeader("X-Real-Ip"))
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
	handle := c.HandlersContext.handlers[c.index]
	c.index++
	handle(c)
}

//GetURLQuery get the params in url (Eg . /?q=)
func (c *Context) GetURLQuery(query string) string {
	return c.Request.URL.Query().Get(query)
}
