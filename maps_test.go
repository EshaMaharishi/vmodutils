package vmodutils

import (
	"testing"

	"go.viam.com/test"
)

func TestGetIntFromMap(t *testing.T) {
	_, ok := GetIntFromMap(map[string]interface{}{}, "x")
	test.That(t, ok, test.ShouldBeFalse)

	_, ok = GetIntFromMap(map[string]interface{}{"x": "Asd"}, "x")
	test.That(t, ok, test.ShouldBeFalse)

	i, ok := GetIntFromMap(map[string]interface{}{"x": 5}, "x")
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, i, test.ShouldEqual, 5)

	i, ok = GetIntFromMap(map[string]interface{}{"x": 5.1}, "x")
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, i, test.ShouldEqual, 5)

}

func TestGetInt64FromMap(t *testing.T) {
	_, ok := GetInt64FromMap(map[string]interface{}{}, "x")
	test.That(t, ok, test.ShouldBeFalse)

	_, ok = GetInt64FromMap(map[string]interface{}{"x": "Asd"}, "x")
	test.That(t, ok, test.ShouldBeFalse)

	i, ok := GetInt64FromMap(map[string]interface{}{"x": 5}, "x")
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, i, test.ShouldEqual, 5)

	i, ok = GetInt64FromMap(map[string]interface{}{"x": 5.1}, "x")
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, i, test.ShouldEqual, 5)

}
