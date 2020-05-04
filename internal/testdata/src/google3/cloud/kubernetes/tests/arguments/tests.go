// Package arguments contains test-cases for testing PII leak detection when sources are introduced via arguments.
package arguments

import (
	"google3/base/go/log"
	"google3/net/proto2/go/proto"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

func testSourceFromParamByReference(c *ipb.Cluster) {
	log.Infof("Source in the parameter %v", c) // want "a source has reached a sink"
}

func testSourceMethodFromParamByReference(c *ipb.Cluster) {
	log.Infof("Source in the parameter %v", c.GetClusterHash())
}

func testSourceFromParamByReferenceInfo(c *ipb.Cluster) {
	log.Info(c) // want "a source has reached a sink"
}

func testSourceFromParamByValue(c ipb.Cluster) {
	log.Infof("Source in the parameter %v", c) // want "a source has reached a sink"
}

func testUpdatedSource(c *ipb.Cluster) {
	c.MasterAuth.Password = proto.String("password")
	log.Infof("Updated cluster %v", c) // want "a source has reached a sink"
}

func testSourceFromAPointerCopy(c *ipb.Cluster) {
	cp := c
	log.Infof("Pointer copy of the cluster %v", cp) // want "a source has reached a sink"
}
