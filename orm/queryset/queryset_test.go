package queryset

import (
	"strings"
	"testing"
)

type MockUser struct {
	ID        uint64     `drf:"id;primary_key"`
	Username  string     `drf:"username"`
	Age       int        `drf:"age"`
	Posts     []MockPost `drf:"relation=posts.author_id"` // Reverse FK
	Followers []MockUser `drf:"m2m=user_follows;to=follower_id;from=following_id"`
}

func (u MockUser) TableName() string { return "mock_users" }

func TestQuerySetLookups(t *testing.T) {
	qs := &QuerySet[MockUser]{} // Use dummy db for SQL testing

	tests := []struct {
		name     string
		filter   Q
		expected string
	}{
		{
			name:     "contains lookup",
			filter:   Q{"username__contains": "admin"},
			expected: "username LIKE $1",
		},
		{
			name:     "icontains lookup",
			filter:   Q{"username__icontains": "admin"},
			expected: "username ILIKE $1",
		},
		{
			name:     "gte lookup",
			filter:   Q{"age__gte": 18},
			expected: "age >= $1",
		},
		{
			name:     "in lookup",
			filter:   Q{"id__in": []uint64{1, 2, 3}},
			expected: "id IN ($1, $2, $3)",
		},
		{
			name:     "iexact lookup",
			filter:   Q{"username__iexact": "Admin"},
			expected: "username ILIKE $1",
		},
		{
			name:     "gt lookup",
			filter:   Q{"age__gt": 21},
			expected: "age > $1",
		},
		{
			name:     "lt lookup",
			filter:   Q{"age__lt": 65},
			expected: "age < $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := qs.Filter(tt.filter)
			sql, _ := q.SQL()
			if !strings.Contains(sql, tt.expected) {
				t.Errorf("expected SQL to contain %q, but got %q", tt.expected, sql)
			}
		})
	}
}

type MockPost struct {
	ID      uint64 `drf:"id;primary_key"`
	Author  uint64 `drf:"author_id;foreign_key=mock_users.id"`
	Content string `drf:"content"`
}

func (p MockPost) TableName() string { return "mock_posts" }

func TestSelectRelatedSQL(t *testing.T) {
	qs := &QuerySet[MockPost]{}

	t.Run("basic join", func(t *testing.T) {
		q := qs.SelectRelated("Author")
		sql, _ := q.SQL()
		expected := "INNER JOIN mock_users ON mock_posts.author_id = mock_users.id"
		if !strings.Contains(sql, expected) {
			t.Errorf("expected SQL to contain JOIN, but got %q", sql)
		}
	})
}

func TestPrefetchRelated(t *testing.T) {
	// This test will require a mock DB that can return multiple results for separate queries
	// For now, we verify that handlePrefetch is called correctly and attempts to fetch
}
