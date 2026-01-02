package urls

import (
	"fmt"
	"net/http"
	"strconv"
)

// Params provides convenient access to URL parameters
type Params struct {
	values map[string]interface{}
}

// Get returns a parameter as a string
func (p *Params) Get(name string) string {
	if p.values == nil {
		return ""
	}

	val, ok := p.values[name]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%v", val)
}

// GetInt returns a parameter as an int
func (p *Params) GetInt(name string) (int, error) {
	strVal := p.Get(name)
	if strVal == "" {
		return 0, fmt.Errorf("parameter %s not found", name)
	}

	return strconv.Atoi(strVal)
}

// GetUint returns a parameter as a uint64
func (p *Params) GetUint(name string) (uint64, error) {
	strVal := p.Get(name)
	if strVal == "" {
		return 0, fmt.Errorf("parameter %s not found", name)
	}

	return strconv.ParseUint(strVal, 10, 64)
}

// GetParams retrieves URL parameters from the request context
func GetParams(r *http.Request) *Params {
	params, ok := r.Context().Value(urlParamsKey).(map[string]interface{})
	if !ok {
		// Return empty params instead of nil
		return &Params{values: make(map[string]interface{})}
	}

	return &Params{values: params}
}
