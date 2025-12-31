package fields

import (
	"strconv"
	"strings"
)

// ParseTag parses the 'drf' struct tag into FieldOptions
func ParseTag(tag string) *FieldOptions {
	opts := &FieldOptions{
		Editable: true,
	}

	if tag == "" {
		return opts
	}

	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		key := kv[0]
		val := ""
		if len(kv) > 1 {
			val = kv[1]
		}

		switch key {
		case "primary_key":
			opts.PrimaryKey = true
		case "unique":
			opts.Unique = true
		case "null":
			opts.Null = true
		case "blank":
			opts.Blank = true
		case "index", "db_index":
			opts.Index = true
		case "default":
			opts.Default = val
		case "max_length":
			opts.MaxLength, _ = strconv.Atoi(val)
		case "decimal":
			opts.Decimal = val
		case "type":
			opts.Type = val
		case "choices":
			opts.Choices = strings.Split(val, ",")
		case "validators":
			opts.Validators = strings.Split(val, ",")
		case "auto_now":
			opts.AutoNow = true
		case "auto_now_add":
			opts.AutoNowAdd = true
		case "db_column":
			opts.DBColumn = val
		case "editable":
			opts.Editable, _ = strconv.ParseBool(val)
		case "help_text":
			opts.HelpText = val
		}
	}

	return opts
}
