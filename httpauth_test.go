package vmodutils

import (
	"testing"

	"go.viam.com/test"
)

func TestHTTPAuthHeader(t *testing.T) {
	res := HTTPAuthHeader("BoatAuth", "abc", "123", "sdlk12qwd")

	test.That(t, res, test.ShouldEqual, "BoatAuth robot_id=\"abc\", api_key_id=\"123\", api_key_hash=\"955e1e866beac69a50c4799c63a9934e6a0e0ce545fe0c66ddb1b7d1efe401c3\"")
}
