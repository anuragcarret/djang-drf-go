package views

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Context provides request context to views
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         map[string]interface{}
	Query          url.Values
	Data           map[string]interface{}
}

// ParseRequest parses the request body based on Content-Type
func (c *Context) ParseRequest() error {
	ct := c.Request.Header.Get("Content-Type")
	if ct == "application/json" {
		if c.Request.Body == nil {
			return nil
		}
		decoder := json.NewDecoder(c.Request.Body)
		return decoder.Decode(&c.Data)
	}
	// TODO: Support other content types (form-data, etc.)
	return nil
}
