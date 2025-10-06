package touch

import (
	"testing"

	"go.viam.com/rdk/pointcloud"
	"go.viam.com/test"
)

func TestCluster1(t *testing.T) {
	in, err := pointcloud.NewFromFile("data/glass1.pcd", "")
	test.That(t, err, test.ShouldBeNil)

	clusters, err := Cluster(in, 30, 20, 100)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, len(clusters), test.ShouldEqual, 1)
}

func BenchmarkCluster1(t *testing.B) {
	in, err := pointcloud.NewFromFile("data/glass1.pcd", "")
	test.That(t, err, test.ShouldBeNil)

	t.ResetTimer()
	for range t.N {
		clusters, err := Cluster(in, 30, 20, 100)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, len(clusters), test.ShouldEqual, 1)
	}
}
