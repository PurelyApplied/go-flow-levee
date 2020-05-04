// Package dominance contains test-cases for testing PII sanitization.
package dominance

import (
	"time"

	"google3/base/go/log"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"

	"google3/cloud/kubernetes/engine/common/sanitize"
)

func testSanitizedSourceDoesNotTriggerFinding(c *ipb.Cluster) {
	sanitized := sanitize.MakeSafeForLogs(c)
	log.Infof("Sanitized cluster %v", sanitized)
}

func testNotGuaranteedSanitization(c *ipb.Cluster) {
	p := c
	if time.Now().Weekday() == time.Monday {
		p = sanitize.MakeSafeForLogs(c).(*ipb.Cluster)
	}
	log.Infof("Sometimes sanitized cluster: %v", p) // want "a source has reached a sink"
}
