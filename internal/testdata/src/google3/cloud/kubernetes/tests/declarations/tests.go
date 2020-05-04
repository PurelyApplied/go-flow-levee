// Package declarations contains test-cases for testing PII leak detection when sources are introduced via declarations.
package declarations

import (
	"google3/base/go/log"
	"google3/net/proto2/go/proto"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

func testSourceDeclaredInBody() {
	c := &ipb.Cluster{}
	log.Infof("%v", c) // want "a source has reached a sink"

	m := &ipb.MasterAuth{Password: proto.String("password")}
	log.Infof("%v", m) // want "a source has reached a sink"

	h := ipb.ClusterHealth{Status: 1}
	log.Infof("%v", h)
}

func testSourceViaClosure() func() {
	c := &ipb.Cluster{}
	return func() {
		log.Infof("Creating a cluster from closure %v", c) // want "a source has reached a sink"
	}
}
