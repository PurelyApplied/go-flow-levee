// Package errors contains test-cases for testing PII leak detection when sources are attached to errors.
package errors

import (
	"errors"
	"fmt"

	"google3/base/go/log"
	"google3/cloud/kubernetes/engine/common/sanitize"
	"google3/net/proto2/go/proto"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

func testAttachSourceToAnError(c *ipb.Cluster) error {
	return fmt.Errorf("failed to create cluster %v", c) // want "a source has reached a sink"
}

func testErrorNoPII(c *ipb.Cluster) error {
	err := errors.New("some error")
	return fmt.Errorf("failed to do something %v", err)
}

func doSomethingWithPII(c *ipb.Cluster) error {
	return fmt.Errorf("failed to do something with cluster %v", c) // want "a source has reached a sink"
}

func testErrorPII(c *ipb.Cluster) error {
	// This function can't sanitize PII prior to calling doSomething since PII may be needed there.
	// At the same time, "err" is linked to PII and may still have PII.
	// The approach here is to check if vararg that are linked to PII are protos - can be sanitized.
	// In other words, if this function can sanitize the returned value then a finding
	// is warranted. Since we can't sanitize an error, then no finding should be issued in this func.
	err := doSomethingWithPII(c)
	return fmt.Errorf("failed to do something %v", err)
}

func testErrorPIIAndNonPII(c *ipb.Cluster) error {
	err := doSomethingWithPII(c)
	// err is tainted, but is not sanitizable.
	// ipb.ClusterHealth is a proto, and therefore sanitizable, but it is not PII.
	return fmt.Errorf("failed to do something %v, non PII %v", err, &ipb.ClusterHealth{})
}

func testSanitizedNewVarNoPII(ctx string, c *ipb.Cluster, e string) {
	newProtoMessage := sanitize.MakeSafeForLogs(c)
	log.Errorf("Null log param. Cluster: %s, HealthStatus: %v", newProtoMessage, e)
}

func testSanitizedNewDeclaredVarNoPII(ctx string, c *ipb.Cluster, e string) {
	var newProtoMessage proto.Message
	newProtoMessage = sanitize.MakeSafeForLogs(c)
	log.Errorf("Null log param. Cluster: %s, HealthStatus: %v", newProtoMessage, e)
}

func testRecastWithIntermediate(ctx string, c *ipb.Cluster, e string) {
	newProtoMessage := sanitize.MakeSafeForLogs(c)
	newCluster := newProtoMessage.(*ipb.Cluster)
	log.Errorf("Null log param. Cluster: %s, HealthStatus: %v", newCluster, e)
}

func testRecastWithoutIntermediate(ctx string, c *ipb.Cluster, e string) {
	newCluster := sanitize.MakeSafeForLogs(c).(*ipb.Cluster)
	log.Errorf("Null log param. Cluster: %s, HealthStatus: %v", newCluster, e)
}
