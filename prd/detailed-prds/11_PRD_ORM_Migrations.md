# PRD: ORM Migrations

> **Module:** `orm/migrations`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.db.migrations`

---

## 1. Overview

### 1.1 Purpose

The Migrations System manages database schema evolution:

- **Schema Versioning**: Track database changes over time
- **Auto-Generation**: Generate migrations from model changes
- **Rollback Support**: Undo migrations safely
- **Multi-Database**: Apply migrations per database

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| Python migration files | Go migration files |
| `RunPython` operations | Go function operations |
| `makemigrations` command | `makemigrations` command |
| `migrate` command | `migrate` command |

---

## 2. Core Concepts

### 2.1 Migration Structure

```go
// Migration represents a schema change
type Migration struct {
    // App this migration belongs to
    App string
    
    // Name unique identifier (e.g., "0001_initial")
    Name string
    
    // Dependencies list of migrations that must run first
    Dependencies []string
    
    // Operations to perform
    Operations []Operation
    
    // Reversible if this migration can be rolled back
    Reversible bool
}
```

### 2.2 Migration File

```go
// users/migrations/0001_initial.go
package migrations

import "github.com/yourusername/django-drf-go/orm/migrations"

func init() {
    migrations.Register(&migrations.Migration{
        App:  "users",
        Name: "0001_initial",
        Dependencies: []string{},
        Operations: []migrations.Operation{
            migrations.CreateModel{
                Name: "User",
                Fields: []migrations.Field{
                    {Name: "id", Type: "bigserial", Options: "PRIMARY KEY"},
                    {Name: "email", Type: "varchar(255)", Options: "NOT NULL UNIQUE"},
                    {Name: "username", Type: "varchar(150)", Options: "NOT NULL UNIQUE"},
                    {Name: "password", Type: "varchar(128)", Options: "NOT NULL"},
                    {Name: "is_active", Type: "boolean", Options: "DEFAULT true"},
                    {Name: "created_at", Type: "timestamp", Options: "DEFAULT now()"},
                },
                Indexes: []migrations.Index{
                    {Fields: []string{"email"}},
                    {Fields: []string{"username", "is_active"}},
                },
            },
        },
    })
}
```

### 2.3 Operations

```go
// CreateModel creates a new table
type CreateModel struct {
    Name        string
    Fields      []Field
    Indexes     []Index
    Constraints []Constraint
}

// DeleteModel drops a table
type DeleteModel struct {
    Name string
}

// AddField adds a column
type AddField struct {
    Model   string
    Field   Field
}

// RemoveField drops a column
type RemoveField struct {
    Model     string
    FieldName string
}

// AlterField modifies a column
type AlterField struct {
    Model     string
    FieldName string
    Field     Field
}

// RenameField renames a column
type RenameField struct {
    Model   string
    OldName string
    NewName string
}

// AddIndex creates an index
type AddIndex struct {
    Model string
    Index Index
}

// RemoveIndex drops an index
type RemoveIndex struct {
    Model     string
    IndexName string
}

// AddConstraint adds a constraint
type AddConstraint struct {
    Model      string
    Constraint Constraint
}

// RemoveConstraint drops a constraint
type RemoveConstraint struct {
    Model          string
    ConstraintName string
}

// RunSQL executes raw SQL
type RunSQL struct {
    SQL        string
    ReverseSQL string  // For rollback
}

// RunGo executes Go code
type RunGo struct {
    Forward func(db *DB) error
    Reverse func(db *DB) error
}
```

---

## 3. API Specification

### 3.1 Migration Commands

#### makemigrations

```bash
# Generate migrations for all apps
./manage makemigrations

# Generate for specific app
./manage makemigrations users

# Dry run (show without creating)
./manage makemigrations --dry-run

# Create empty migration
./manage makemigrations users --empty --name add_custom_logic

# Merge conflicting migrations
./manage makemigrations --merge
```

#### migrate

```bash
# Apply all migrations
./manage migrate

# Apply specific app
./manage migrate users

# Migrate to specific point
./manage migrate users 0003

# Rollback all for an app
./manage migrate users zero

# Show migration plan
./manage migrate --plan

# Fake migration (mark as applied without running)
./manage migrate --fake

# Fake initial (fake if tables exist)
./manage migrate --fake-initial
```

#### showmigrations

```bash
# Show all migrations and their status
./manage showmigrations

# Show for specific app
./manage showmigrations users

# Output:
# users
#  [X] 0001_initial
#  [X] 0002_add_profile
#  [ ] 0003_add_avatar
```

#### sqlmigrate

```bash
# Show SQL for a migration
./manage sqlmigrate users 0001

# Output:
# CREATE TABLE users_user (
#     id BIGSERIAL PRIMARY KEY,
#     email VARCHAR(255) NOT NULL UNIQUE,
#     ...
# );
```

### 3.2 Programmatic API

```go
// Apply migrations
err := migrations.Migrate("default", nil)  // All apps

// Apply specific app
err := migrations.Migrate("default", &MigrateOptions{
    App: "users",
})

// Rollback
err := migrations.Migrate("default", &MigrateOptions{
    App:    "users",
    Target: "0002",  // Rollback to 0002
})

// Check pending
pending, err := migrations.Pending("default")
for _, m := range pending {
    fmt.Printf("%s.%s\n", m.App, m.Name)
}
```

---

## 4. Auto-Detection

### 4.1 Model Changes Detected

| Change | Operation Generated |
|--------|-------------------|
| New model | `CreateModel` |
| Deleted model | `DeleteModel` |
| New field | `AddField` |
| Removed field | `RemoveField` |
| Changed field type | `AlterField` |
| Renamed field | `RenameField` (with prompt) |
| New index | `AddIndex` |
| Removed index | `RemoveIndex` |
| New unique constraint | `AddConstraint` |

### 4.2 Detection Process

```go
// Detector compares current models to migration history
type Detector struct {
    currentState  SchemaState  // From registered models
    appliedState  SchemaState  // From migration history
}

// Detect finds differences
func (d *Detector) Detect() []Operation

// Usage in makemigrations:
detector := migrations.NewDetector(app)
operations := detector.Detect()
if len(operations) > 0 {
    migration := migrations.NewMigration(app, operations)
    migration.Write()
}
```

---

## 5. Django Comparison

### 5.1 Migration File

**Django:**
```python
# users/migrations/0002_add_avatar.py
from django.db import migrations, models

class Migration(migrations.Migration):
    dependencies = [
        ('users', '0001_initial'),
    ]
    
    operations = [
        migrations.AddField(
            model_name='user',
            name='avatar',
            field=models.ImageField(upload_to='avatars/', blank=True),
        ),
    ]
```

**Django-DRF-Go:**
```go
// users/migrations/0002_add_avatar.go
package migrations

import m "github.com/yourusername/django-drf-go/orm/migrations"

func init() {
    m.Register(&m.Migration{
        App:  "users",
        Name: "0002_add_avatar",
        Dependencies: []string{"users.0001_initial"},
        Operations: []m.Operation{
            m.AddField{
                Model: "User",
                Field: m.Field{
                    Name:    "avatar",
                    Type:    "varchar(255)",
                    Options: "DEFAULT ''",
                },
            },
        },
    })
}
```

### 5.2 Data Migration

**Django:**
```python
def forward(apps, schema_editor):
    User = apps.get_model('users', 'User')
    for user in User.objects.all():
        user.full_name = f"{user.first_name} {user.last_name}"
        user.save()

class Migration(migrations.Migration):
    operations = [
        migrations.RunPython(forward, migrations.RunPython.noop),
    ]
```

**Django-DRF-Go:**
```go
m.RunGo{
    Forward: func(db *orm.DB) error {
        users, err := db.Raw("SELECT id, first_name, last_name FROM users").All()
        if err != nil {
            return err
        }
        for _, u := range users {
            fullName := fmt.Sprintf("%s %s", u["first_name"], u["last_name"])
            _, err = db.Exec("UPDATE users SET full_name = ? WHERE id = ?", fullName, u["id"])
            if err != nil {
                return err
            }
        }
        return nil
    },
    Reverse: func(db *orm.DB) error {
        _, err := db.Exec("UPDATE users SET full_name = ''")
        return err
    },
}
```

---

## 6. Migration State

### 6.1 Applied Migrations Table

```sql
CREATE TABLE django_migrations (
    id BIGSERIAL PRIMARY KEY,
    app VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    applied TIMESTAMP DEFAULT now()
);
```

### 6.2 Checking State

```go
// AppliedMigrations returns applied migration names
func AppliedMigrations(db string) ([]string, error)

// IsApplied checks if specific migration is applied
func IsApplied(db, app, name string) (bool, error)

// MarkApplied records migration as applied (for --fake)
func MarkApplied(db, app, name string) error

// MarkUnapplied removes migration record
func MarkUnapplied(db, app, name string) error
```

---

## 7. Squashing Migrations

```bash
# Squash migrations into one
./manage squashmigrations users 0001 0010

# Creates: 0001_squashed_0001_0010.go
```

```go
// Squashed migration
&m.Migration{
    App:  "users",
    Name: "0001_squashed_0001_0010",
    Replaces: []string{
        "0001_initial",
        "0002_add_email",
        // ...
        "0010_add_avatar",
    },
    Operations: []m.Operation{
        // Combined operations
    },
}
```

---

## 8. Multi-Database Support

```go
// Apply to specific database
./manage migrate --database=analytics

// In code
err := migrations.Migrate("analytics", nil)

// Database router for apps
type DatabaseRouter struct{}

func (r *DatabaseRouter) DBForMigrate(app string) string {
    if app == "analytics" {
        return "analytics"
    }
    return "default"
}
```

---

## 9. Conflict Resolution

### 9.1 Detecting Conflicts

```bash
./manage makemigrations

# Output:
# Conflicting migrations detected:
#   users.0003_add_email
#   users.0003_add_phone
# Run 'manage makemigrations --merge' to fix.
```

### 9.2 Merging

```bash
./manage makemigrations --merge
```

```go
// Generated merge migration
&m.Migration{
    App:  "users",
    Name: "0004_merge_0003_add_email_0003_add_phone",
    Dependencies: []string{
        "users.0003_add_email",
        "users.0003_add_phone",
    },
    Operations: []m.Operation{}, // Empty merge
}
```

---

## 10. Testing Migrations

```go
func TestMigrations(t *testing.T) {
    // Create test database
    testDB := migrations.CreateTestDB()
    defer migrations.DropTestDB(testDB)
    
    // Apply all migrations
    err := migrations.Migrate(testDB, nil)
    assert.NoError(t, err)
    
    // Rollback
    err = migrations.Migrate(testDB, &MigrateOptions{
        App:    "users",
        Target: "zero",
    })
    assert.NoError(t, err)
    
    // Re-apply
    err = migrations.Migrate(testDB, nil)
    assert.NoError(t, err)
}
```

---

## 11. Error Handling

| Error | Cause | Resolution |
|-------|-------|------------|
| `ErrCircularDependency` | Migration depends on itself | Fix dependency chain |
| `ErrMissingDependency` | Required migration not found | Create missing migration |
| `ErrIrreversible` | Cannot rollback migration | Use `RunSQL` with `ReverseSQL` |
| `ErrPartialApply` | Migration failed mid-way | Manual fix or restore backup |

---

## 12. Best Practices

1. **Always backup** before running migrations in production
2. **Test migrations** on staging first
3. **Keep migrations small** - one logical change per migration
4. **Never edit applied migrations** - create new ones
5. **Use data migrations** for data transformations
6. **Provide reverse SQL** for raw SQL operations

---

## 13. Related PRDs

- [06_PRD_Core_Management_Commands.md](./06_PRD_Core_Management_Commands.md) - Migration commands
- [07_PRD_ORM_Models.md](./07_PRD_ORM_Models.md) - Model definitions
- [08_PRD_ORM_Fields.md](./08_PRD_ORM_Fields.md) - Field types
