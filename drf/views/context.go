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
	User           interface{}
}

// ParseRequest parses the request body based on Content-Type
func (c *Context) ParseRequest() error {
	ct := c.Request.Header.Get("Content-Type")
	if c.Request.Body == nil {
		return nil
	}

	if ct == "application/json" {
		decoder := json.NewDecoder(c.Request.Body)
		return decoder.Decode(&c.Data)
	}
	// TODO: Support other content types (form-data, etc.)
	return nil
}

// URLParam retrieves a URL path parameter (e.g., /items/:id)
// For now, falls back to query params until router integration is complete
func (c *Context) URLParam(key string) string {
	// TODO: Integrate with router to extract path params
	// For now, use query params as fallback
	return c.Query.Get(key)
}

// Bind unmarshals the parsed data into v
func (c *Context) Bind(v interface{}) error {
	// Since data is already in a map, we can marshal it back to json
	// and unmarshal into the struct, or use a mapstructure library.
	// For simplicity in this framework, we'll use json as a bridge.
	data, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
