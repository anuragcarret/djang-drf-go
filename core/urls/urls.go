package urls

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// contextKey for URL parameters
type contextKey struct{}

var urlParamsKey = contextKey{}

// URLPattern defines a route
type URLPattern struct {
	Pattern  string
	Regex    *regexp.Regexp
	Params   []paramInfo
	Handler  http.Handler
	Name     string
	Methods  []string
	Includes *Router
}

type paramInfo struct {
	Name      string
	Converter string
}

// Router manages URL patterns
type Router struct {
	patterns  []*URLPattern
	namespace string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		patterns: make([]*URLPattern, 0),
	}
}

// Register adds a list of patterns to the router
func (r *Router) Register(patterns ...*URLPattern) {
	for _, p := range patterns {
		p.compile()
		r.patterns = append(r.patterns, p)
	}
}

// Path registers a route matching all methods
func (r *Router) Path(pattern string, handler http.Handler, name string) *Router {
	return r.Methods(nil, pattern, handler, name)
}

// Get registers a GET route
func (r *Router) Get(pattern string, handler http.Handler, name string) *Router {
	return r.Methods([]string{"GET"}, pattern, handler, name)
}

// Post registers a POST route
func (r *Router) Post(pattern string, handler http.Handler, name string) *Router {
	return r.Methods([]string{"POST"}, pattern, handler, name)
}

// Methods registers a route with specific methods
func (r *Router) Methods(methods []string, pattern string, handler http.Handler, name string) *Router {
	p := &URLPattern{
		Pattern: pattern,
		Handler: handler,
		Name:    name,
		Methods: methods,
	}
	p.compile()
	r.patterns = append(r.patterns, p)
	return r
}

// Include adds all patterns from another router under a prefix
func (r *Router) Include(prefix string, router *Router, namespace string) *Router {
	p := &URLPattern{
		Pattern:  prefix,
		Includes: router,
		Name:     namespace,
	}
	router.namespace = namespace
	r.patterns = append(r.patterns, p)
	return r
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	allowedMethods := make(map[string]bool)
	pathMatched := false

	for _, p := range r.patterns {
		if p.Includes != nil {
			if strings.HasPrefix(path, p.Pattern) {
				// Strip prefix and delegate
				http.StripPrefix(p.Pattern, p.Includes).ServeHTTP(w, req)
				return
			}
			continue
		}

		match := p.Regex.FindStringSubmatch(path)
		if match != nil {
			pathMatched = true

			// Check methods
			if p.Methods != nil {
				matchesMethod := false
				for _, m := range p.Methods {
					allowedMethods[m] = true
					if m == req.Method {
						matchesMethod = true
						break
					}
				}
				if !matchesMethod {
					continue
				}
			}

			// Extract params
			params := make(map[string]interface{})
			for i, info := range p.Params {
				val := match[i+1]
				conv, _ := GetConverter(info.Converter)
				goVal, _ := conv.ToGo(val)
				params[info.Name] = goVal
			}

			ctx := context.WithValue(req.Context(), urlParamsKey, params)
			p.Handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}

	if pathMatched {
		if req.Method == "OPTIONS" {
			methods := make([]string, 0, len(allowedMethods))
			for m := range allowedMethods {
				methods = append(methods, m)
			}
			w.Header().Set("Allow", strings.Join(methods, ", "))
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, req)
}

// --- Pattern Helpers ---

// MakePath creates a standalone URLPattern (useful for defining lists)
func MakePath(pattern string, handler http.Handler, name string) *URLPattern {
	return &URLPattern{Pattern: pattern, Handler: handler, Name: name}
}

// MakeGet creates a GET URLPattern
func MakeGet(pattern string, handler http.Handler, name string) *URLPattern {
	return &URLPattern{Pattern: pattern, Handler: handler, Name: name, Methods: []string{"GET"}}
}

// MakePost creates a POST URLPattern
func MakePost(pattern string, handler http.Handler, name string) *URLPattern {
	return &URLPattern{Pattern: pattern, Handler: handler, Name: name, Methods: []string{"POST"}}
}

// Reverse generates a URL from a name and parameters
func (r *Router) Reverse(name string, args ...interface{}) (string, error) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) > 1 {
		// Namespaced reverse
		for _, p := range r.patterns {
			if p.Includes != nil && p.Name == parts[0] {
				subURL, err := p.Includes.Reverse(parts[1], args...)
				if err != nil {
					return "", err
				}
				return p.Pattern + subURL, nil
			}
		}
		return "", fmt.Errorf("namespace %s not found", parts[0])
	}

	for _, p := range r.patterns {
		if p.Name == name {
			if p.Includes != nil {
				return p.Pattern, nil
			}

			url := p.Pattern
			for i, info := range p.Params {
				if i >= len(args) {
					return "", fmt.Errorf("missing argument for %s", info.Name)
				}
				conv, _ := GetConverter(info.Converter)
				strVal, _ := conv.ToURL(args[i])
				placeholder := fmt.Sprintf("{%s:%s}", info.Name, info.Converter)
				url = strings.Replace(url, placeholder, strVal, 1)
			}
			return url, nil
		}
	}

	return "", fmt.Errorf("url name %s not found", name)
}

// URLParam returns a single parameter value
func URLParam[T any](r *http.Request, name string) (T, error) {
	var zero T
	params, ok := r.Context().Value(urlParamsKey).(map[string]interface{})
	if !ok {
		return zero, fmt.Errorf("no url params found")
	}
	val, ok := params[name]
	if !ok {
		return zero, fmt.Errorf("param %s not found", name)
	}
	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("param %s is not of type %T", name, zero)
	}
	return typed, nil
}

// compile prepares the regex for a pattern
func (p *URLPattern) compile() {
	pattern := p.Pattern
	re := regexp.MustCompile(`\{([^}:]+):([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)

	p.Params = make([]paramInfo, 0, len(matches))
	regexStr := "^" + pattern + "$"

	for _, match := range matches {
		name := match[1]
		convName := match[2]
		p.Params = append(p.Params, paramInfo{Name: name, Converter: convName})

		conv, ok := GetConverter(convName)
		if !ok {
			// Default to string converter if not found
			conv, _ = GetConverter("str")
		}

		placeholder := fmt.Sprintf("{%s:%s}", name, convName)
		regexStr = strings.Replace(regexStr, placeholder, "("+conv.Regex()+")", 1)
	}

	p.Regex = regexp.MustCompile(regexStr)
}
