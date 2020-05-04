// Package proto contains test-cases for testing PII leak detection when sources are manipulated via proto library.
package proto

import (
	"bytes"

	"google3/base/go/log"
	"google3/net/proto2/go/proto"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"

	"google3/cloud/kubernetes/engine/common/sanitize"
)

func testMarshalProto(c *ipb.Cluster) {
	var buf bytes.Buffer
	proto.MarshalText(&buf, c)
	log.Info(buf) // want "a source has reached a sink"
}

func testMarshalStringProto(c *ipb.Cluster) {
	msg := proto.MarshalTextString(c)
	log.Info(msg) // want "a source has reached a sink"
}

func testMarshalStringSanitizedProto(c *ipb.Cluster) {
	msg := proto.MarshalTextString(sanitize.MakeSafeForLogs(c))
	log.Infof(msg)
}

func testClonedSource(c *ipb.Cluster) {
	clone := proto.Clone(c)
	log.Infof("%v", clone) // want "a source has reached a sink"
}
