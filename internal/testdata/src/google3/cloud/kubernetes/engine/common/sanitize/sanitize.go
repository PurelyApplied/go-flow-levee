// Package sanitize fakes cloud/kubernetes/engine/common/sanitize.
package sanitize

import (
	"google3/net/proto2/go/proto"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

// MakeSafeForLogs fakes MakeSafeForLogs.
func MakeSafeForLogs(in proto.Message) proto.Message {
	return &ipb.Cluster{}
}
