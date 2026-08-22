package gameserver

import (
	"errors"
	"testing"
)

func TestDeploymentBoundaryBindsOneOriginToOneTrustedProxyHop(t *testing.T) {
	origins, hops, err := deploymentBoundary(CompositionConfig{PublicOrigin: "https://play.example.test", TrustedProxyHops: 1})
	if err != nil || hops != 1 || len(origins) != 1 || origins[0] != "https://play.example.test" {
		t.Fatalf("origins=%v hops=%d err=%v", origins, hops, err)
	}
	devOrigins, devHops, err := deploymentBoundary(CompositionConfig{})
	if err != nil || devOrigins != nil || devHops != 0 {
		t.Fatalf("development origins=%v hops=%d err=%v", devOrigins, devHops, err)
	}
	for _, invalid := range []CompositionConfig{
		{TrustedProxyHops: 1},
		{PublicOrigin: "https://play.example.test", TrustedProxyHops: 0},
		{PublicOrigin: "https://play.example.test", TrustedProxyHops: 2},
		{PublicOrigin: "http://play.example.test", TrustedProxyHops: 1},
		{PublicOrigin: "https://play.example.test/path", TrustedProxyHops: 1},
	} {
		if _, _, err := deploymentBoundary(invalid); !errors.Is(err, ErrComposition) {
			t.Fatalf("invalid deployment boundary accepted: %+v err=%v", invalid, err)
		}
	}
}
