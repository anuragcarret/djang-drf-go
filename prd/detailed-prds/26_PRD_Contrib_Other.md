# PRD: Contrib Other Modules

> **Module:** `contrib/`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** Various `django.contrib` modules

---

## 1. ContentTypes

### 1.1 Purpose

Generic model references for polymorphic relations.

```go
// ContentType represents a model type
type ContentType struct {
    orm.Model
    AppLabel  string `drf:"max_length=100"`
    Model     string `drf:"max_length=100"`
}

// Get content type for model
ct := contenttypes.GetForModel(&Article{})
// ct.AppLabel = "articles", ct.Model = "article"

// Get model from content type
model := ct.ModelClass()
```

### 1.2 Generic Foreign Key

```go
type Comment struct {
    orm.Model
    ContentTypeID uint64
    ObjectID      uint64
    Content       string
}

// Get related object
func (c *Comment) ContentObject() interface{} {
    ct, _ := contenttypes.Get(c.ContentTypeID)
    return ct.GetObject(c.ObjectID)
}
```

---

## 2. StaticFiles

### 2.1 Purpose

Serve static files (CSS, JS, images).

```go
// Development: serve from app directories
router.Static("/static/", staticfiles.Finder{})

// Production: collect to single directory
./manage collectstatic

// Settings
{
    "static": {
        "url": "/static/",
        "root": "./staticfiles/",
        "dirs": ["./static/"]
    }
}
```

### 2.2 Template Helper

```go
// In templates
<link rel="stylesheet" href="{{ static "css/style.css" }}">
// Output: /static/css/style.css
```

---

## 3. Messages

### 3.1 Purpose

Flash messages displayed after redirect.

```go
// Add message
messages.Success(r, "Profile updated successfully!")
messages.Error(r, "Please correct the errors below.")
messages.Warning(r, "Your session will expire soon.")
messages.Info(r, "New features are available.")

// In template
{{ range .Messages }}
    <div class="alert alert-{{ .Level }}">{{ .Text }}</div>
{{ end }}
```

### 3.2 Message Levels

```go
const (
    DEBUG   Level = 10
    INFO    Level = 20  
    SUCCESS Level = 25
    WARNING Level = 30
    ERROR   Level = 40
)
```

---

## 4. Humanize

### 4.1 Template Filters

```go
// Number formatting
{{ 1000000 | intcomma }}        // "1,000,000"
{{ 1.5 | floatformat:2 }}       // "1.50"

// Time
{{ .CreatedAt | naturaltime }}  // "2 hours ago"
{{ .CreatedAt | naturalday }}   // "yesterday"

// Size
{{ 1048576 | filesizeformat }}  // "1.0 MB"

// Text
{{ 5 | ordinal }}               // "5th"
{{ 1 | apnumber }}              // "one"
```

---

## 5. Sitemaps

### 5.1 Purpose

Generate XML sitemaps for SEO.

```go
type ArticleSitemap struct {
    sitemaps.Sitemap
}

func (s *ArticleSitemap) Items() []interface{} {
    articles, _ := orm.Objects[Article]().
        Filter(orm.Q{"is_published": true}).
        All()
    return articles
}

func (s *ArticleSitemap) Location(item interface{}) string {
    return fmt.Sprintf("/articles/%s/", item.(*Article).Slug)
}

func (s *ArticleSitemap) Lastmod(item interface{}) time.Time {
    return item.(*Article).UpdatedAt
}

// Register
sitemaps.Register("articles", &ArticleSitemap{})

// URL
router.Get("/sitemap.xml", sitemaps.Index)
```

---

## 6. Syndication (RSS/Atom)

```go
type ArticleFeed struct {
    feeds.Feed
}

func (f *ArticleFeed) Title() string {
    return "Latest Articles"
}

func (f *ArticleFeed) Items() []interface{} {
    articles, _ := orm.Objects[Article]().
        OrderBy("-published_at").
        Limit(20).
        All()
    return articles
}

// URL
router.Get("/feed/", feeds.RSS(&ArticleFeed{}))
router.Get("/atom/", feeds.Atom(&ArticleFeed{}))
```

---

## 7. Related PRDs

- [24_PRD_Contrib_Auth.md](./24_PRD_Contrib_Auth.md) - Auth
- [25_PRD_Contrib_Sessions.md](./25_PRD_Contrib_Sessions.md) - Sessions
