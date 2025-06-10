package vmodutils

import (
	"testing"

	"go.viam.com/test"
)

func TestState(t *testing.T) {
	s := StringState{}
	test.That(t, s.String(), test.ShouldEqual, "")

	s.Push("a")
	test.That(t, s.String(), test.ShouldEqual, "a")

	s.Push("b")
	test.That(t, s.String(), test.ShouldEqual, "a,b")

	s.Pop()
	test.That(t, s.String(), test.ShouldEqual, "a")

	err := s.CheckEmpty()
	test.That(t, err, test.ShouldNotBeNil)

	s.Pop()
	test.That(t, s.String(), test.ShouldEqual, "")

	s.Pop()
	test.That(t, s.String(), test.ShouldEqual, "")

}
