package mint

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
)

//constant
const (
	emptyString     = ""
	PayloadKey      = "PayLoad"
	contentType     = "Content-Type"
	contentEncoding = "Content-Encoding"
)

var (
	newLine             = []byte{'\n'}
	jsonContentType     = []string{"application/json; charset=utf-8"}
	gzipContentEncoding = []string{"gzip"}
)

//Context provides context for whole request/response cycle
//It helps to pass variable from one middlware to another
type Context struct {
	HandlerContext *HandlerContext
	Req            *http.Request
	Res            http.ResponseWriter
	params         map[string]string
	index          int
	status         int
	size           int
	errors         []error
	query          url.Values
}

func (app *Mint) newContext() *Context {
	return new(Context)
}

//Reset resets the value the context
func (c *Context) Reset() {
	c.HandlerContext = nil
	c.Req = nil
	c.Res = nil
	c.status = 0
	c.size = 0
	c.errors = c.errors[0:0]
	c.index = 0
	c.params = nil
	c.query = nil
}
func newContextPool(app *Mint) func() interface{} {
	return func() interface{} {
		return app.newContext()
	}
}

//GetHeader returns request header
//shortcut for c.Rewq
func (c *Context) GetHeader(key string) string {
	return c.Req.Header.Get(key)
}

//Status set http status code
func (c *Context) Status(status int) {
	c.status = status
	c.Res.WriteHeader(status)
}

func (c *Context) writeContentType(values []string) {
	c.SetHeader(contentType, values)
}

//SetHeader #
func (c *Context) SetHeader(key string, value []string) {
	header := c.Res.Header()
	if val := header[key]; len(val) == 0 {
		header[key] = value
	}
}

//AddHeader add header to the response
func (c *Context) AddHeader(key string, value string) {
	c.Res.Header().Add(key, value)
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

func (c *Context) jsonMarshal(reponse interface{}) []byte {
	jsonContentByte, err := json.Marshal(reponse)
	if err != nil {
		c.Errors(err)
	}
	return jsonContentByte
}

//CJSON writes compressed json response
func (c *Context) compressedJSON(code int, response interface{}) {
	// create header
	c.SetHeader(contentEncoding, gzipContentEncoding)
	// Gzip data
	c.Status(code)
	gz := c.HandlerContext.Mint.gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(c.Res)
	jsonContentByte := c.jsonMarshal(response)
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
	c.HandlerContext.Mint.gzipWriterPool.Put(gz)
}

//JSON #
func (c *Context) JSON(code int, response interface{}) {
	c.SetHeader(contentType, jsonContentType)
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
func (c *Context) uncompressedJSON(code int, response interface{}) {
	c.Status(code)
	jsonContentByte := c.jsonMarshal(response)
	size, err := c.Res.Write(jsonContentByte)
	if err != nil {
		c.Errors(err)
	}
	c.setSize(size)
	size, err = c.Res.Write(newLine)
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
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Req.RemoteAddr)); err == nil {
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

//QueryArray #
func (c *Context) QueryArray(key string) ([]string, bool) {
	if c.query == nil {
		c.query = c.Req.URL.Query()
	}
	if values, ok := c.query[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

//Query #
func (c *Context) Query(key string) (string, bool) {
	queryArray, ok := c.QueryArray(key)
	if ok {
		return queryArray[0], ok
	}
	return "", ok
}

//DefaultQuery #
func (c *Context) DefaultQuery(key string, defaultv string) (string, bool) {
	value, ok := c.Query(key)
	if ok {
		return value, ok
	}
	return defaultv, ok
}

//QueryValues #
func (c *Context) QueryValues() url.Values {
	return c.query
}

//Get #
func (c *Context) Get(key interface{}) interface{} {
	return c.Req.Context().Value(key)
}

//DefaultGet #
func (c *Context) DefaultGet(key interface{}, defaultv interface{}) interface{} {
	value := c.Get(key)
	if value == nil {
		return defaultv
	}
	return value
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
	c.Req = c.Req.WithContext(context.WithValue(c.Req.Context(), key, val))
}

//Param #
func (c *Context) Param(key string) (string, bool) {
	value, ok := c.params[key]
	if ok {
		return value, ok
	}
	return "", ok
}

//DefaultParam #
func (c *Context) DefaultParam(key string, defaultv string) string {
	value, ok := c.params[key]
	if ok {
		return value
	}
	return defaultv
}

//ParamMap #
func (c *Context) ParamMap() map[string]string {
	return c.params
}

//MintParam #
func (c *Context) MintParam(key string) (interface{}, bool) {
	value, ok := c.HandlerContext.Mint.Get(key)
	return value, ok
}

//SetMintParam #
func (c *Context) SetMintParam(key string, value interface{}) {
	c.HandlerContext.Mint.Set(key, value)
}

// URI gets request uri
func (c *Context) URI() string {
	return c.Req.RequestURI
}
