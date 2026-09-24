package spec

import (
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Expander resolves ${VAR} / ${VAR:default} references in non-credential
// string fields against the tool's own environment at generate time.
type Expander struct {
	// Lookup resolves a variable name (os.LookupEnv shape). nil disables
	// expansion entirely: Expand becomes a no-op.
	Lookup func(string) (string, bool)
	// Warn reports one unset, defaultless variable so a typo cannot vanish
	// silently. Called once per occurrence; nil suppresses the report but the
	// verbatim passthrough still happens.
	Warn func(format string, a ...any)
}

// expandRE matches only the braced form: a bare $VAR is literal text (rule 4).
// The default group is everything up to the closing brace, so a default may
// not itself contain '}'.
var expandRE = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:[^}]*)?\}`)

var yamlNodeType = reflect.TypeOf(yaml.Node{})

// Expand rewrites every expandable string field of e and wfs in place. A field
// is left untouched when it is tagged `expand:"no"` (credentials, rule 5) or is
// a *yaml.Node / yaml.Node verbatim-passthrough field (rule 6); every other
// string field is rewritten only when it actually contains "${".
func Expand(ex Expander, e *Env, wfs []Workflow) {
	if ex.Lookup == nil {
		return
	}
	if e != nil {
		expandValue(ex, reflect.ValueOf(e).Elem())
	}
	for i := range wfs {
		expandValue(ex, reflect.ValueOf(&wfs[i]).Elem())
	}
}

// expandValue walks an addressable value, rewriting string fields in place.
func expandValue(ex Expander, v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		expandValue(ex, v.Elem())
	case reflect.Struct:
		if v.Type() == yamlNodeType {
			return // rule 6: verbatim passthrough, never expanded
		}
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if !f.IsExported() || f.Tag.Get("expand") == "no" {
				continue
			}
			expandValue(ex, v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			expandValue(ex, v.Index(i))
		}
	case reflect.Map:
		expandMap(ex, v)
	case reflect.String:
		if v.CanSet() {
			if s := v.String(); strings.Contains(s, "${") {
				v.SetString(expandString(ex, s))
			}
		}
	}
}

// expandMap read-modify-writes each entry: a map value is not addressable, so
// it is copied into an addressable local, walked, and written back.
func expandMap(ex Expander, v reflect.Value) {
	if v.IsNil() {
		return
	}
	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key()
		tmp := reflect.New(v.Type().Elem()).Elem()
		tmp.Set(iter.Value())
		expandValue(ex, tmp)
		v.SetMapIndex(k, tmp)
	}
}

// expandString substitutes every ${VAR} / ${VAR:default} occurrence in s.
// Unset with a default uses the default; unset with no default passes the
// occurrence through verbatim and reports it via ex.Warn (rule 3).
func expandString(ex Expander, s string) string {
	return expandRE.ReplaceAllStringFunc(s, func(m string) string {
		sub := expandRE.FindStringSubmatch(m)
		name, rawDefault := sub[1], sub[2]
		if v, ok := ex.Lookup(name); ok {
			return v
		}
		if rawDefault != "" {
			return strings.TrimPrefix(rawDefault, ":")
		}
		if ex.Warn != nil {
			ex.Warn("variable %s is not set and has no default; leaving %q unexpanded", name, m)
		}
		return m
	})
}
