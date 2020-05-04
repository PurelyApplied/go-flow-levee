// Package deploy contains test-cases for testing PII leak detection when sources emitted via error_desc.go GKEErrorDescriptor
package deploy

import (
	"google3/cloud/kubernetes/engine/common/errdesc"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

func buildP(clusterP *ipb.Cluster, clusterV ipb.Cluster) {
	errdesc.InternalError.WithMsg("InternalError in cluster %v", clusterP) // want "a source has reached a sink"
	errdesc.AlreadyExists.WithMsg("AlreadyExists in cluster %v", clusterV) // want "a source has reached a sink"
}
