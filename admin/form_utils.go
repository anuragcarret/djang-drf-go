package admin

// Form field structures for template rendering
type FormField struct {
	Name      string
	Label     string
	Widget    string // "text", "email", "number", "checkbox", "textarea", "date", "datetime", "select"
	Value     interface{}
	Required  bool
	ReadOnly  bool
	HelpText  string
	Error     string
	MaxLength int
	Step      string
	Choices   []FormChoice
}

type FormChoice struct {
	Value string
	Label string
}

type FormFieldSet struct {
	Name        string
	Description string
	Fields      []FormField
}
