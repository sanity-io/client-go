package sanity

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Param creates a named parameter.
func Param(name string, value interface{}) Parameter {
	return Parameter{Name: name, Value: value}
}

// Parameter is a query parameter.
type Parameter struct {
	// Name is the name of the parameter, which can be used in the query as $name.
	Name string

	// Value is the value of the parameter. The value must be JSON-serializable.
	Value interface{}
}

func (p *Parameter) build(uv url.Values) error {
	b, err := json.Marshal(p.Value)
	if err != nil {
		return fmt.Errorf("could not marshal parameter %q to JSON: %w", p.Name, err)
	}
	uv["$"+p.Name] = []string{string(b)}
	return nil
}
