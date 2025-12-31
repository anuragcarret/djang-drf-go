# PRD: Core Management Commands

> **Module:** `core/management`  
> **Version:** 1.0.0  
> **Status:** Draft  
> **Django Equivalent:** `django.core.management`

---

## 1. Overview

### 1.1 Purpose

The Management Commands System provides CLI tools for project administration:

- **Built-in Commands**: Common operations (migrations, server, shell, etc.)
- **Custom Commands**: App-specific CLI tools
- **Argument Parsing**: Type-safe flag handling
- **Interactive Mode**: User prompts and confirmation

### 1.2 Design Philosophy

| Django Approach | Go-Idiomatic Approach |
|-----------------|----------------------|
| `python manage.py <command>` | Binary with subcommands |
| `BaseCommand` class | `Command` interface |
| `argparse` arguments | Struct-based flags |
| `call_command()` | `Execute()` function |

---

## 2. Core Concepts

### 2.1 Command Interface

```go
// Command defines a management command
type Command interface {
    // Name returns the command name (used in CLI)
    Name() string
    
    // Help returns short description
    Help() string
    
    // Run executes the command
    Run(ctx context.Context, args []string) error
}

// CommandWithFlags adds flag parsing
type CommandWithFlags interface {
    Command
    
    // Flags returns flag definitions
    Flags() *FlagSet
}
```

### 2.2 Command Registry

```go
// Registry manages available commands
type Registry struct {
    commands map[string]Command
}

var Commands = NewRegistry()

// Register adds a command
func (r *Registry) Register(cmd Command)

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, error)

// All returns all registered commands
func (r *Registry) All() []Command
```

### 2.3 Built-in Commands

| Command | Description |
|---------|-------------|
| `runserver` | Start development server |
| `migrate` | Apply database migrations |
| `makemigrations` | Generate migration files |
| `shell` | Interactive Go console |
| `createsuperuser` | Create admin user |
| `collectstatic` | Gather static files |
| `check` | Validate project configuration |
| `dbshell` | Open database CLI |
| `help` | Show available commands |

---

## 3. API Specification

### 3.1 Defining Commands

#### Simple Command

```go
package commands

import (
    "context"
    "github.com/yourusername/django-drf-go/core/management"
)

type HelloCommand struct{}

func (c *HelloCommand) Name() string { return "hello" }
func (c *HelloCommand) Help() string { return "Say hello" }

func (c *HelloCommand) Run(ctx context.Context, args []string) error {
    management.Stdout.Println("Hello, World!")
    return nil
}

func init() {
    management.Commands.Register(&HelloCommand{})
}
```

#### Command with Flags

```go
type MigrateCommand struct {
    App      string `flag:"app" short:"a" help:"App name to migrate"`
    Fake     bool   `flag:"fake" help:"Mark migrations as applied"`
    Database string `flag:"database" short:"d" default:"default" help:"Database alias"`
}

func (c *MigrateCommand) Name() string { return "migrate" }
func (c *MigrateCommand) Help() string { return "Apply database migrations" }

func (c *MigrateCommand) Flags() *management.FlagSet {
    return management.BindFlags(c)
}

func (c *MigrateCommand) Run(ctx context.Context, args []string) error {
    management.Stdout.Printf("Migrating %s on database %s\n", c.App, c.Database)
    
    if c.Fake {
        management.Stdout.Println("Faking migrations...")
    }
    
    return migrations.RunMigrations(c.Database, c.App, c.Fake)
}
```

### 3.2 Running Commands

#### CLI Entry Point

```go
// manage.go
package main

import (
    "os"
    "github.com/yourusername/django-drf-go/core/management"
    
    // Import apps to register their commands
    _ "myproject/users"
    _ "myproject/orders"
)

func main() {
    if err := management.Execute(os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

#### Programmatic Execution

```go
// Call command from code
err := management.CallCommand("migrate", "--app", "users", "--fake")

// With context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := management.CallCommandWithContext(ctx, "migrate", "--app", "users")
```

### 3.3 Flag System

```go
// FlagSet defines command arguments
type FlagSet struct {
    flags map[string]*Flag
}

type Flag struct {
    Name     string
    Short    string
    Help     string
    Default  interface{}
    Required bool
    Type     FlagType
}

// Bind flags from struct tags
func BindFlags(v interface{}) *FlagSet
```

---

## 4. Built-in Commands Detail

### 4.1 runserver

```go
// Usage: manage runserver [host:port]
type RunServerCommand struct {
    Address    string `flag:"addr" default:":8000" help:"Address to bind"`
    NoReload   bool   `flag:"no-reload" help:"Disable auto-reload"`
    NoColor    bool   `flag:"no-color" help:"Disable colored output"`
}

// Examples:
// manage runserver
// manage runserver :9000
// manage runserver 0.0.0.0:8000 --no-reload
```

### 4.2 migrate

```go
// Usage: manage migrate [app] [migration]
type MigrateCommand struct {
    App      string `flag:"app" help:"App label"`
    Fake     bool   `flag:"fake" help:"Mark as applied without running"`
    FakeInitial bool `flag:"fake-initial" help:"Fake if tables exist"`
    Plan     bool   `flag:"plan" help:"Show migration plan"`
    Database string `flag:"database" default:"default" help:"Database alias"`
}

// Examples:
// manage migrate                    # Apply all
// manage migrate users              # Apply users migrations
// manage migrate users 0003         # Up to specific migration
// manage migrate users zero         # Rollback all users
// manage migrate --plan             # Show plan
```

### 4.3 makemigrations

```go
// Usage: manage makemigrations [app...]
type MakeMigrationsCommand struct {
    Apps     []string `positional:"apps" help:"Apps to create migrations for"`
    DryRun   bool     `flag:"dry-run" help:"Don't write files"`
    Merge    bool     `flag:"merge" help:"Resolve migration conflicts"`
    Empty    bool     `flag:"empty" help:"Create empty migration"`
    Name     string   `flag:"name" short:"n" help:"Migration name"`
}

// Examples:
// manage makemigrations
// manage makemigrations users
// manage makemigrations --name add_email_field
// manage makemigrations --merge
```

### 4.4 createsuperuser

```go
// Usage: manage createsuperuser
type CreateSuperuserCommand struct {
    Username    string `flag:"username" help:"Admin username"`
    Email       string `flag:"email" help:"Admin email"`
    NoInput     bool   `flag:"no-input" help:"Non-interactive mode"`
    Database    string `flag:"database" default:"default"`
}

// Interactive prompts if flags not provided
```

### 4.5 shell

```go
// Usage: manage shell
type ShellCommand struct {
    Interface string `flag:"interface" short:"i" default:"default" help:"Shell type"`
}

// Starts interactive REPL with preloaded models
```

### 4.6 check

```go
// Usage: manage check
type CheckCommand struct {
    Tags       []string `flag:"tag" short:"t" help:"Check specific tags"`
    ListTags   bool     `flag:"list-tags" help:"List available tags"`
    FailLevel  string   `flag:"fail-level" default:"ERROR" help:"Fail threshold"`
    Deploy     bool     `flag:"deploy" help:"Run deployment checks"`
}

// Examples:
// manage check
// manage check --deploy
// manage check --tag security --tag database
```

### 4.7 dbshell

```go
// Usage: manage dbshell
type DBShellCommand struct {
    Database string `flag:"database" default:"default" help:"Database alias"`
}

// Opens psql, mysql, sqlite3 depending on database engine
```

### 4.8 collectstatic

```go
// Usage: manage collectstatic
type CollectStaticCommand struct {
    NoInput    bool     `flag:"no-input" help:"Skip confirmation"`
    Clear      bool     `flag:"clear" short:"c" help:"Clear existing files"`
    Link       bool     `flag:"link" short:"l" help:"Create symlinks"`
    DryRun     bool     `flag:"dry-run" help:"Don't actually copy"`
    IgnorePatterns []string `flag:"ignore" short:"i" help:"Ignore patterns"`
}
```

---

## 5. Django Comparison

### 5.1 Command Definition

**Django:**
```python
# myapp/management/commands/mycommand.py
from django.core.management.base import BaseCommand

class Command(BaseCommand):
    help = 'Does something useful'
    
    def add_arguments(self, parser):
        parser.add_argument('name', type=str)
        parser.add_argument('--force', action='store_true')
    
    def handle(self, *args, **options):
        name = options['name']
        if options['force']:
            self.stdout.write('Forcing...')
        self.stdout.write(f'Hello, {name}!')
```

**Django-DRF-Go:**
```go
// myapp/commands/mycommand.go
package commands

import "github.com/yourusername/django-drf-go/core/management"

type MyCommand struct {
    Name  string `positional:"name" required:"true"`
    Force bool   `flag:"force" help:"Force operation"`
}

func (c *MyCommand) Name() string { return "mycommand" }
func (c *MyCommand) Help() string { return "Does something useful" }

func (c *MyCommand) Flags() *management.FlagSet {
    return management.BindFlags(c)
}

func (c *MyCommand) Run(ctx context.Context, args []string) error {
    if c.Force {
        management.Stdout.Println("Forcing...")
    }
    management.Stdout.Printf("Hello, %s!\n", c.Name)
    return nil
}

func init() {
    management.Commands.Register(&MyCommand{})
}
```

### 5.2 Running Commands

**Django:**
```bash
python manage.py migrate --database=default
python manage.py createsuperuser --username=admin
```

**Django-DRF-Go:**
```bash
./manage migrate --database=default
./manage createsuperuser --username=admin
```

### 5.3 Calling Commands Programmatically

**Django:**
```python
from django.core.management import call_command

call_command('migrate', '--database=default')
call_command('loaddata', 'fixtures.json', verbosity=0)
```

**Django-DRF-Go:**
```go
import "github.com/yourusername/django-drf-go/core/management"

management.CallCommand("migrate", "--database=default")
management.CallCommand("loaddata", "fixtures.json", "--verbosity=0")
```

---

## 6. Interactive Features

### 6.1 User Prompts

```go
// Input prompts
name, err := management.Prompt("Enter your name: ")

// Password (hidden input)
password, err := management.PromptPassword("Password: ")

// Confirmation
confirmed := management.Confirm("Are you sure?")

// Choice
choice := management.Choice("Select database:", []string{"postgres", "mysql", "sqlite"})
```

### 6.2 Progress Indicators

```go
// Progress bar
bar := management.NewProgressBar(total)
for i := 0; i < total; i++ {
    // work...
    bar.Increment()
}
bar.Finish()

// Spinner
spinner := management.NewSpinner("Loading...")
spinner.Start()
// work...
spinner.Stop()
```

### 6.3 Styled Output

```go
// Colored output
management.Stdout.Success("Operation completed successfully")
management.Stdout.Warning("This might take a while")
management.Stdout.Error("Something went wrong")
management.Stdout.Info("Processing 100 items...")

// Styled text
management.Stdout.Bold("Important message")
management.Stdout.Dim("Less important")
```

---

## 7. Command Groups

### 7.1 Sub-Commands

```go
// db group with subcommands
type DBCommand struct{}

func (c *DBCommand) Name() string { return "db" }
func (c *DBCommand) Help() string { return "Database operations" }

func (c *DBCommand) Subcommands() []Command {
    return []Command{
        &DBShellCommand{},
        &DBDumpCommand{},
        &DBLoadCommand{},
    }
}

// Usage:
// manage db shell
// manage db dump --output=backup.sql
// manage db load backup.sql
```

### 7.2 App Commands

```go
// Commands scoped to apps
type AppCommand struct {
    baseApp string
}

func (c *AppCommand) Name() string {
    return fmt.Sprintf("%s.mycommand", c.baseApp)
}

// Usage: manage users.cleanup
```

---

## 8. Testing Commands

### 8.1 Test Helpers

```go
import "github.com/yourusername/django-drf-go/core/management/testing"

func TestMigrateCommand(t *testing.T) {
    // Capture output
    out := testing.CaptureOutput(func() {
        management.CallCommand("migrate", "--plan")
    })
    
    assert.Contains(t, out, "0001_initial")
}
```

### 8.2 Mock Input

```go
func TestInteractiveCommand(t *testing.T) {
    // Provide mock input
    testing.MockInput("admin", "admin@example.com", "password123")
    
    err := management.CallCommand("createsuperuser")
    assert.NoError(t, err)
}
```

---

## 9. Error Handling

### 9.1 Command Errors

```go
// CommandError for non-critical errors
type CommandError struct {
    Message string
    Code    int
}

func (c *MyCommand) Run(ctx context.Context, args []string) error {
    if invalid {
        return &management.CommandError{
            Message: "Invalid input",
            Code:    1,
        }
    }
    return nil
}
```

### 9.2 Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Configuration error |

---

## 10. Configuration

### 10.1 Management Settings

```go
type ManagementConfig struct {
    // DefaultDatabase for commands
    DefaultDatabase string `json:"default_database" default:"default"`
    
    // Color output enabled
    Color bool `json:"color" default:"true"`
    
    // Verbosity level (0-3)
    Verbosity int `json:"verbosity" default:"1"`
}
```

---

## 11. Implementation Notes

### 11.1 Dependencies

| Package | Purpose |
|---------|---------|
| `os/exec` | Running external commands |
| `bufio` | Interactive input |
| `term` | Terminal detection |
| `color` | Colored output |

### 11.2 Build Integration

```go
// Embed version info at build time
// go build -ldflags "-X main.Version=1.0.0"

var Version = "dev"

type VersionCommand struct{}

func (c *VersionCommand) Run(ctx context.Context, args []string) error {
    management.Stdout.Printf("django-drf-go version %s\n", Version)
    return nil
}
```

---

## 12. Related PRDs

- [01_PRD_Core_Apps_System.md](./01_PRD_Core_Apps_System.md) - App-specific commands
- [02_PRD_Core_Settings.md](./02_PRD_Core_Settings.md) - Settings for commands
- [11_PRD_ORM_Migrations.md](./11_PRD_ORM_Migrations.md) - Migration commands
- [21_PRD_Admin_Dashboard.md](./21_PRD_Admin_Dashboard.md) - Admin commands
