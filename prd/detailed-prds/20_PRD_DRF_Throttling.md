# PRD: DRF Throttling

> **Module:** `drf/throttling`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **DRF Equivalent:** `rest_framework.throttling`

---

## 1. Overview

### 1.1 Purpose

Rate limiting to protect APIs from abuse:

- **Request Rate Limiting**: Limit requests per time period
- **User vs Anonymous**: Different limits for authenticated users
- **Scoped Throttles**: Per-view or per-action limits
- **Burst Protection**: Allow bursts with sustained limits

---

## 2. Built-in Throttles

### 2.1 AnonRateThrottle

```go
drf.AnonRateThrottle("100/day")  // Anonymous users

// Identifies by IP address
```

### 2.2 UserRateThrottle

```go
drf.UserRateThrottle("1000/day")  // Authenticated users

// Identifies by user ID
```

### 2.3 ScopedRateThrottle

```go
// Different rates for different endpoints
{
    "throttle_rates": {
        "uploads": "5/hour",
        "contacts": "100/day",
        "burst": "10/minute"
    }
}

type UploadViewSet struct {
    drf.ModelViewSet[Upload]
}

func (v *UploadViewSet) GetThrottles() []drf.Throttle {
    return []drf.Throttle{drf.ScopedRateThrottle("uploads")}
}
```

---

## 3. Rate Format

```
"<number>/<period>"

Periods:
- second, sec, s
- minute, min, m
- hour, h
- day, d

Examples:
- "100/day"
- "60/minute"
- "10/second"
```

---

## 4. Configuration

```go
// Global settings
{
    "rest_framework": {
        "default_throttle_classes": ["AnonRateThrottle", "UserRateThrottle"],
        "throttle_rates": {
            "anon": "100/day",
            "user": "1000/day"
        }
    }
}

// Per-ViewSet override
func (v *APIViewSet) GetThrottles() []drf.Throttle {
    return []drf.Throttle{
        drf.UserRateThrottle("100/hour"),
    }
}
```

---

## 5. Custom Throttles

```go
type PremiumUserThrottle struct {
    drf.SimpleRateThrottle
}

func (t *PremiumUserThrottle) GetRate(c *drf.Context) string {
    if c.User != nil && c.User.IsPremium {
        return "10000/day"
    }
    return "100/day"
}
```

---

## 6. Throttle Headers

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 3600
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1704067200
```

---

## 7. DRF Comparison

**DRF:**
```python
REST_FRAMEWORK = {
    'DEFAULT_THROTTLE_CLASSES': [
        'rest_framework.throttling.AnonRateThrottle',
        'rest_framework.throttling.UserRateThrottle'
    ],
    'DEFAULT_THROTTLE_RATES': {
        'anon': '100/day',
        'user': '1000/day'
    }
}
```

**Django-DRF-Go:**
```json
{
    "rest_framework": {
        "default_throttle_classes": ["AnonRateThrottle", "UserRateThrottle"],
        "throttle_rates": {
            "anon": "100/day",
            "user": "1000/day"
        }
    }
}
```

---

## 8. Related PRDs

- [02_PRD_Core_Settings.md](./02_PRD_Core_Settings.md) - Settings
- [15_PRD_DRF_Views_ViewSets.md](./15_PRD_DRF_Views_ViewSets.md) - ViewSets
