package touch

import (
	"testing"

	"go.viam.com/rdk/referenceframe"
	"go.viam.com/test"
)

func TestHashString(t *testing.T) {
	test.That(t, HashString(""), test.ShouldEqual, HashString(""))
	test.That(t, HashString("eliot"), test.ShouldEqual, HashString("eliot"))

	test.That(t, HashString(""), test.ShouldNotEqual, HashString("a"))
	test.That(t, HashString(""), test.ShouldNotEqual, HashString("eliot"))
}

func TestHashInputs(t *testing.T) {
	a := []referenceframe.Input{{5}, {7}}
	b := []referenceframe.Input{{5.0001}, {7.0001}}
	c := []referenceframe.Input{{5.1}, {7.1}}
	d := []referenceframe.Input{{7}, {5}}

	test.That(t, HashInputs(a), test.ShouldEqual, HashInputs(a))
	test.That(t, HashInputs(a), test.ShouldEqual, HashInputs(b))
	test.That(t, HashInputs(a), test.ShouldNotEqual, HashInputs(c))
	test.That(t, HashInputs(a), test.ShouldNotEqual, HashInputs(d))
}
