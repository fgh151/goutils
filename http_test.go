package sdk

import (
	"net"
	"testing"
)

func TestNetworks(t *testing.T) {

	_, local, _ := net.ParseCIDR("172.31.0.8/16")
	_, remote, _ := net.ParseCIDR("172.31.0.7/16")

	if intersect(local, remote) == false {
		t.Fatal("Wrong addresses intersect")
	}
}
