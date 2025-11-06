package capability

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type AnyUnmarshaler interface {
	UnmarshalAny(data map[string]any) error
}

// FiltersDecodeHook returns a mapstructure.DecodeHookFunc that builds Filters from raw map[string]interface{}.
func FiltersDecodeHook() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.Map || from.Key().Kind() != reflect.String || from.Elem().Kind() != reflect.Interface {
			return data, nil
		}
		result := reflect.New(to).Interface()
		decoder, ok := result.(AnyUnmarshaler)
		if !ok {
			return data, nil
		}
		mapData, ok := data.(map[string]any)
		if !ok {
			return data, nil
		}
		if err := decoder.UnmarshalAny(mapData); err != nil {
			return nil, err
		}
		return result, nil
	}
}
