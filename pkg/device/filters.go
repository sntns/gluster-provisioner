package device

import (
	"fmt"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

// Filters centralizes the logic for chaining filters (AND/OR).
type Filters struct {
	Items []Filter  `mapstructure:"items"`
	Op    LogicalOp `mapstructure:"op"`
}

// UnmarshalAny builds a Filters object from a mapstructure-unmarshalled map[string]interface{}.
func (f *Filters) UnmarshalAny(raw map[string]any) error {
	op, ok := raw["op"].(string)
	if ok {
		f.Op = LogicalOp(op)
	}
	items, ok := raw["items"].([]interface{})
	if !ok {
		return fmt.Errorf("items must be an array")
	}
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("item must be a map")
		}
		t, ok := m["type"].(string)
		if !ok {
			return fmt.Errorf("filter type missing")
		}
		switch t {
		case "path":
			pf := PathFilter{}
			if pattern, ok := m["pattern"].(string); ok {
				pf.Pattern = pattern
			}
			f.Items = append(f.Items, pf)
		case "filters":
			var nested Filters
			err := nested.UnmarshalAny(m)
			if err != nil {
				return err
			}
			f.Items = append(f.Items, nested)
		default:
			return fmt.Errorf("unknown filter type: %s", t)
		}
	}
	return nil
}

var _ Filter = Filters{}

// Match applies the AND/OR logic to the chained filters.
func (fs Filters) Match(info model.DeviceInfo) bool {
	if len(fs.Items) == 0 {
		return true
	}
	switch fs.Op {
	case OpAnd:
		for _, f := range fs.Items {
			if !f.Match(info) {
				return false
			}
		}
		return true
	case OpOr:
		for _, f := range fs.Items {
			if f.Match(info) {
				return true
			}
		}
		return false
	default:
		// Default to OR behavior
		for _, f := range fs.Items {
			if f.Match(info) {
				return true
			}
		}
		return false
	}
}
