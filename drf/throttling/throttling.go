package throttling

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Throttle defines the interface for all throttle classes
type Throttle interface {
	AllowRequest(r *http.Request, view interface{}) bool
	GetCacheKey(r *http.Request, view interface{}) string
	GetRate() string
	Wait() time.Duration
}

// BaseThrottle provides default throttle behavior
type BaseThrottle struct {
	Rate        string
	Storage     ThrottleStorage
	NumRequests int
	Duration    time.Duration
}

// ThrottleStorage defines the interface for storing rate limit data
type ThrottleStorage interface {
	Get(key string) (int, time.Time, bool)
	Set(key string, count int, expiry time.Time)
	Increment(key string, expiry time.Time) int
	Clear(key string)
}

// InMemoryThrottleStorage implements in-memory storage for rate limits
type InMemoryThrottleStorage struct {
	data map[string]*throttleEntry
	mu   sync.RWMutex
}

type throttleEntry struct {
	count  int
	expiry time.Time
}

func NewInMemoryThrottleStorage() *InMemoryThrottleStorage {
	storage := &InMemoryThrottleStorage{
		data: make(map[string]*throttleEntry),
	}
	// Start cleanup goroutine
	go storage.cleanup()
	return storage
}

func (s *InMemoryThrottleStorage) Get(key string) (int, time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return 0, time.Time{}, false
	}

	// Check if expired
	if time.Now().After(entry.expiry) {
		return 0, time.Time{}, false
	}

	return entry.count, entry.expiry, true
}

func (s *InMemoryThrottleStorage) Set(key string, count int, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &throttleEntry{
		count:  count,
		expiry: expiry,
	}
}

func (s *InMemoryThrottleStorage) Increment(key string, expiry time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists || time.Now().After(entry.expiry) {
		s.data[key] = &throttleEntry{
			count:  1,
			expiry: expiry,
		}
		return 1
	}

	entry.count++
	return entry.count
}

func (s *InMemoryThrottleStorage) Clear(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *InMemoryThrottleStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.data {
			if now.After(entry.expiry) {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}

// ParseRate parses a rate string like "100/hour" into num requests and duration
func ParseRate(rateStr string) (int, time.Duration, error) {
	parts := strings.Split(rateStr, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid rate format: %s", rateStr)
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid rate number: %s", parts[0])
	}

	var duration time.Duration
	switch parts[1] {
	case "second", "sec", "s":
		duration = time.Second
	case "minute", "min", "m":
		duration = time.Minute
	case "hour", "hr", "h":
		duration = time.Hour
	case "day", "d":
		duration = 24 * time.Hour
	default:
		return 0, 0, fmt.Errorf("invalid rate period: %s", parts[1])
	}

	return num, duration, nil
}

// AnonRateThrottle throttles anonymous users by IP address
type AnonRateThrottle struct {
	BaseThrottle
}

func NewAnonRateThrottle(rate string, storage ThrottleStorage) *AnonRateThrottle {
	num, duration, _ := ParseRate(rate)
	return &AnonRateThrottle{
		BaseThrottle: BaseThrottle{
			Rate:        rate,
			Storage:     storage,
			NumRequests: num,
			Duration:    duration,
		},
	}
}

func (t *AnonRateThrottle) AllowRequest(r *http.Request, view interface{}) bool {
	// Check if user is authenticated (skip for authenticated users)
	user := r.Context().Value("user")
	if user != nil {
		return true // Don't throttle authenticated users
	}

	cacheKey := t.GetCacheKey(r, view)
	if cacheKey == "" {
		return true
	}

	expiry := time.Now().Add(t.Duration)
	count := t.Storage.Increment(cacheKey, expiry)

	return count <= t.NumRequests
}

func (t *AnonRateThrottle) GetCacheKey(r *http.Request, view interface{}) string {
	// Use IP address as cache key
	ip := getClientIP(r)
	return fmt.Sprintf("throttle_anon_%s", ip)
}

func (t *AnonRateThrottle) GetRate() string {
	return t.Rate
}

func (t *AnonRateThrottle) Wait() time.Duration {
	return t.Duration
}

// UserRateThrottle throttles authenticated users by user ID
type UserRateThrottle struct {
	BaseThrottle
}

func NewUserRateThrottle(rate string, storage ThrottleStorage) *UserRateThrottle {
	num, duration, _ := ParseRate(rate)
	return &UserRateThrottle{
		BaseThrottle: BaseThrottle{
			Rate:        rate,
			Storage:     storage,
			NumRequests: num,
			Duration:    duration,
		},
	}
}

func (t *UserRateThrottle) AllowRequest(r *http.Request, view interface{}) bool {
	cacheKey := t.GetCacheKey(r, view)
	if cacheKey == "" {
		return true
	}

	expiry := time.Now().Add(t.Duration)
	count := t.Storage.Increment(cacheKey, expiry)

	return count <= t.NumRequests
}

func (t *UserRateThrottle) GetCacheKey(r *http.Request, view interface{}) string {
	user := r.Context().Value("user")
	if user == nil {
		// Fall back to IP for anonymous users
		ip := getClientIP(r)
		return fmt.Sprintf("throttle_user_anon_%s", ip)
	}

	// Get user ID
	type IDGetter interface {
		GetID() interface{}
	}

	if userWithID, ok := user.(IDGetter); ok {
		userID := userWithID.GetID()
		return fmt.Sprintf("throttle_user_%v", userID)
	}

	return ""
}

func (t *UserRateThrottle) GetRate() string {
	return t.Rate
}

func (t *UserRateThrottle) Wait() time.Duration {
	return t.Duration
}

// ScopedRateThrottle provides different rates for different scopes
type ScopedRateThrottle struct {
	BaseThrottle
	Scopes map[string]string // scope -> rate
}

func NewScopedRateThrottle(scopes map[string]string, storage ThrottleStorage) *ScopedRateThrottle {
	return &ScopedRateThrottle{
		BaseThrottle: BaseThrottle{
			Storage: storage,
		},
		Scopes: scopes,
	}
}

func (t *ScopedRateThrottle) AllowRequest(r *http.Request, view interface{}) bool {
	scope := t.getScope(view)
	rate, exists := t.Scopes[scope]
	if !exists {
		return true // No rate limit for this scope
	}

	// Parse rate for this scope
	num, duration, err := ParseRate(rate)
	if err != nil {
		return true
	}

	t.NumRequests = num
	t.Duration = duration
	t.Rate = rate

	cacheKey := t.GetCacheKey(r, view)
	if cacheKey == "" {
		return true
	}

	expiry := time.Now().Add(duration)
	count := t.Storage.Increment(cacheKey, expiry)

	return count <= num
}

func (t *ScopedRateThrottle) GetCacheKey(r *http.Request, view interface{}) string {
	scope := t.getScope(view)
	user := r.Context().Value("user")

	var ident string
	if user == nil {
		ident = getClientIP(r)
	} else {
		type IDGetter interface {
			GetID() interface{}
		}
		if userWithID, ok := user.(IDGetter); ok {
			ident = fmt.Sprintf("%v", userWithID.GetID())
		} else {
			ident = getClientIP(r)
		}
	}

	return fmt.Sprintf("throttle_scope_%s_%s", scope, ident)
}

func (t *ScopedRateThrottle) getScope(view interface{}) string {
	// Try to get scope from view
	type ScopeGetter interface {
		GetThrottleScope() string
	}

	if scopeView, ok := view.(ScopeGetter); ok {
		return scopeView.GetThrottleScope()
	}

	return "default"
}

func (t *ScopedRateThrottle) GetRate() string {
	return t.Rate
}

func (t *ScopedRateThrottle) Wait() time.Duration {
	return t.Duration
}

// Helper to extract client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// CheckThrottles checks all throttles and returns true if request is allowed
func CheckThrottles(r *http.Request, view interface{}, throttles []Throttle) (bool, time.Duration) {
	for _, throttle := range throttles {
		if !throttle.AllowRequest(r, view) {
			return false, throttle.Wait()
		}
	}
	return true, 0
}
