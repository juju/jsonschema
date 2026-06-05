// Copyright 2026 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package jsonschema

import (
	gc "gopkg.in/check.v1"
)

type conformYAMLSuite struct{}

var _ = gc.Suite(&conformYAMLSuite{})

// Primitive / default-case passthrough.

func (s *conformYAMLSuite) TestNilPassthrough(c *gc.C) {
	result, err := conformYAML(nil)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.IsNil)
}

func (s *conformYAMLSuite) TestStringPassthrough(c *gc.C) {
	result, err := conformYAML("hello")
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.Equals, "hello")
}

func (s *conformYAMLSuite) TestIntPassthrough(c *gc.C) {
	result, err := conformYAML(42)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.Equals, 42)
}

func (s *conformYAMLSuite) TestBoolPassthrough(c *gc.C) {
	result, err := conformYAML(true)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.Equals, true)
}

// map[string]any — keys already strings, values recursed.

func (s *conformYAMLSuite) TestStringKeyedMapPassthrough(c *gc.C) {
	input := map[string]any{"key": "value", "num": 1}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, map[string]any{"key": "value", "num": 1})
}

func (s *conformYAMLSuite) TestStringKeyedMapWithNestedAnyMap(c *gc.C) {
	// map[string]any whose value is a map[any]any — the nested map must be
	// converted.
	input := map[string]any{
		"nested": map[any]any{"inner": 99},
	}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, map[string]any{
		"nested": map[string]any{"inner": 99},
	})
}

func (s *conformYAMLSuite) TestStringKeyedMapErrorPropagates(c *gc.C) {
	// A non-string key in a nested map[any]any causes the error to bubble up.
	input := map[string]any{
		"bad": map[any]any{123: "oops"},
	}
	_, err := conformYAML(input)
	c.Assert(err, gc.NotNil)
	c.Check(err.Error(), gc.Equals, "map keyed with non-string value")
}

// map[any]any — keys converted to strings.

func (s *conformYAMLSuite) TestAnyKeyedMapConverted(c *gc.C) {
	input := map[any]any{"a": 1, "b": "two"}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, map[string]any{"a": 1, "b": "two"})
}

func (s *conformYAMLSuite) TestAnyKeyedMapNonStringKeyErrors(c *gc.C) {
	input := map[any]any{42: "value"}
	_, err := conformYAML(input)
	c.Assert(err, gc.NotNil)
	c.Check(err.Error(), gc.Equals, "map keyed with non-string value")
}

func (s *conformYAMLSuite) TestAnyKeyedMapValuesRecursed(c *gc.C) {
	// Values inside map[any]any are themselves recursed when the outer map
	// is re-processed as map[string]any.
	input := map[any]any{
		"outer": map[any]any{"inner": "val"},
	}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, map[string]any{
		"outer": map[string]any{"inner": "val"},
	})
}

// []any — each element recursed.

func (s *conformYAMLSuite) TestSlicePassthrough(c *gc.C) {
	input := []any{1, "two", true}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, []any{1, "two", true})
}

func (s *conformYAMLSuite) TestSliceWithNestedAnyMap(c *gc.C) {
	input := []any{map[any]any{"k": "v"}}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, []any{map[string]any{"k": "v"}})
}

func (s *conformYAMLSuite) TestSliceErrorPropagates(c *gc.C) {
	input := []any{map[any]any{99: "bad"}}
	_, err := conformYAML(input)
	c.Assert(err, gc.NotNil)
	c.Check(err.Error(), gc.Equals, "map keyed with non-string value")
}

// Deeply nested structures.

func (s *conformYAMLSuite) TestDeeplyNestedConversion(c *gc.C) {
	// map[any]any -> []any -> map[any]any
	input := map[any]any{
		"level1": []any{
			map[any]any{
				"level2": map[any]any{"level3": 42},
			},
		},
	}
	result, err := conformYAML(input)
	c.Assert(err, gc.IsNil)
	c.Check(result, gc.DeepEquals, map[string]any{
		"level1": []any{
			map[string]any{
				"level2": map[string]any{"level3": 42},
			},
		},
	})
}

func (s *conformYAMLSuite) TestDeeplyNestedErrorPropagates(c *gc.C) {
	// Error deep inside a nested structure propagates to the top.
	input := map[any]any{
		"level1": []any{
			map[any]any{
				"level2": map[any]any{true: "bad-key"},
			},
		},
	}
	_, err := conformYAML(input)
	c.Assert(err, gc.NotNil)
	c.Check(err.Error(), gc.Equals, "map keyed with non-string value")
}
