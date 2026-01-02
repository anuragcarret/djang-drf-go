package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const migrationTemplate = `package migrations

import (
	"github.com/anuragcarret/djang-drf-go/orm/migrations"
)

func init() {
	migrations.GlobalRegistry.Register("{{.AppLabel}}", &migrations.Migration{
		ID: "{{.ID}}",
		Operations: []migrations.Operation{
			{{range .Operations}}
			{{if eq .Type "CreateTable"}}
			&migrations.CreateTable{
				Name: "{{.Name}}",
				Fields: map[string]string{
					{{range $k, $v := .Fields}}"{{$k}}": "{{$v}}",
					{{end}}
				},
			},
			{{else if eq .Type "AddField"}}
			&migrations.AddField{
				TableName: "{{.TableName}}",
				FieldName: "{{.FieldName}}",
				FieldType: "{{.FieldType}}",
			},
			{{else if eq .Type "AlterField"}}
			&migrations.AlterField{
				TableName: "{{.TableName}}",
				FieldName: "{{.FieldName}}",
				FieldType: "{{.FieldType}}",
			},
			{{else if eq .Type "RunSQL"}}
			&migrations.RunSQL{
				SQL: "{{.SQL}}",
			},
			{{end}}
			{{end}}
		},
	})
}
`

// Writer generates Go source files for migrations
type Writer struct {
	AppLabel  string
	OutputDir string
}

func NewWriter(appLabel, outputDir string) *Writer {
	return &Writer{AppLabel: appLabel, OutputDir: outputDir}
}

func (w *Writer) Write(ops []Operation) (string, error) {
	if len(ops) == 0 {
		return "", fmt.Errorf("no operations to write")
	}

	id := time.Now().Format("20060102_150405")
	filename := filepath.Join(w.OutputDir, id+"_auto.go")

	type opData struct {
		Type      string
		Name      string
		Fields    map[string]string
		TableName string
		FieldName string
		FieldType string
		SQL       string
	}

	data := struct {
		AppLabel   string
		ID         string
		Operations []opData
	}{
		AppLabel: w.AppLabel,
		ID:       id,
	}

	for _, op := range ops {
		switch o := op.(type) {
		case *CreateTable:
			data.Operations = append(data.Operations, opData{
				Type:   "CreateTable",
				Name:   o.Name,
				Fields: o.Fields,
			})
		case *AddField:
			data.Operations = append(data.Operations, opData{
				Type:      "AddField",
				TableName: o.TableName,
				FieldName: o.FieldName,
				FieldType: o.FieldType,
			})
		case *AlterField:
			data.Operations = append(data.Operations, opData{
				Type:      "AlterField",
				TableName: o.TableName,
				FieldName: o.FieldName,
				FieldType: o.FieldType,
			})
		case *RunSQL:
			data.Operations = append(data.Operations, opData{
				Type: "RunSQL",
				SQL:  o.SQL,
			})
		}
	}

	if err := os.MkdirAll(w.OutputDir, 0755); err != nil {
		return "", err
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	tmpl := template.Must(template.New("migration").Parse(migrationTemplate))
	if err := tmpl.Execute(f, data); err != nil {
		return "", err
	}

	return filename, nil
}
