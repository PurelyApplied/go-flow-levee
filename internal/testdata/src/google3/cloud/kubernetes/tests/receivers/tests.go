// Package receivers contains test-cases for testing PII leak detection when sources are introduced via receivers.
package receivers

import (
	"google3/base/go/log"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

type clusterBuilder struct {
	clusterP *ipb.Cluster
	clusterV ipb.Cluster
}

func (b *clusterBuilder) buildP() {
	log.Infof("Building cluster %v", b.clusterP) // want "a source has reached a sink"
	log.Infof("Building cluster %v", b.clusterV) // want "a source has reached a sink"
}

func (b clusterBuilder) buildV() {
	log.Infof("Building cluster %v", b.clusterP) // want "a source has reached a sink"
	log.Infof("Building cluster %v", b.clusterV) // want "a source has reached a sink"
}
