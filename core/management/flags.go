package management

import (
	"flag"
	"reflect"
	"strconv"
)

// FlagSet wraps standard flag.FlagSet with struct binding
type FlagSet struct {
	fs *flag.FlagSet
	v  interface{}
}

// BindFlags creates a FlagSet from struct tags
func BindFlags(v interface{}) *FlagSet {
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()

	fs := flag.NewFlagSet(rt.Name(), flag.ContinueOnError)

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		flagName := field.Tag.Get("flag")
		if flagName == "" {
			continue
		}

		help := field.Tag.Get("help")
		def := field.Tag.Get("default")

		switch field.Type.Kind() {
		case reflect.String:
			fs.StringVar(fv.Addr().Interface().(*string), flagName, def, help)
		case reflect.Bool:
			defBool, _ := strconv.ParseBool(def)
			fs.BoolVar(fv.Addr().Interface().(*bool), flagName, defBool, help)
		case reflect.Int:
			defInt, _ := strconv.Atoi(def)
			fs.IntVar(fv.Addr().Interface().(*int), flagName, defInt, help)
		}
	}

	return &FlagSet{fs: fs, v: v}
}

// Parse arguments into the bound struct
func (f *FlagSet) Parse(args []string) error {
	return f.fs.Parse(args)
}

// Args returns remaining non-flag arguments
func (f *FlagSet) Args() []string {
	return f.fs.Args()
}
