// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package jsonschema

import "errors"

// conformYAML ensures all keys of any nested maps are strings. This is
// necessary because YAML unmarshals map[any]any in nested maps, which
// cannot be serialised by json or bson. Also handles []any.
// cf. gopkg.in/juju/charm.v4/actions.go cleanse
//
// Adapted from github.com/juju/utils ConformYAML.
func conformYAML(input any) (any, error) {
	switch typedInput := input.(type) {

	case map[string]any:
		newMap := make(map[string]any)
		for key, value := range typedInput {
			newValue, err := conformYAML(value)
			if err != nil {
				return nil, err
			}
			newMap[key] = newValue
		}
		return newMap, nil

	case map[any]any:
		newMap := make(map[string]any)
		for key, value := range typedInput {
			typedKey, ok := key.(string)
			if !ok {
				return nil, errors.New("map keyed with non-string value")
			}
			newMap[typedKey] = value
		}
		return conformYAML(newMap)

	case []any:
		newSlice := make([]any, len(typedInput))
		for i, sliceValue := range typedInput {
			newSliceValue, err := conformYAML(sliceValue)
			if err != nil {
				return nil, err
			}
			newSlice[i] = newSliceValue
		}
		return newSlice, nil

	default:
		return input, nil
	}
}
