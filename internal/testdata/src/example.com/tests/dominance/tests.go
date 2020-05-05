// Package dominance contains test-cases for testing PII sanitization.
package dominance

import (
	"time"

	"example.com/core"
)

func testSanitizedSourceDoesNotTriggerFinding(c *core.Source) {
	sanitized := core.Sanitize(c)
	core.Sinkf("Sanitized %v", sanitized)
}

func testSanitizedSourceDoesNotTriggerFindingWhenTypeAsserted(c *core.Source) {
	sanitized := core.Sanitize(c)[0].(*core.Source)
	core.Sinkf("Sanitized %v", sanitized)
}

func testNotGuaranteedSanitization(c *core.Source) {
	p := c
	if time.Now().Weekday() == time.Monday {
		p = core.Sanitize(c)[0].(*core.Source)
	}
	core.Sinkf("Sometimes sanitized: %v", p) // want "a source has reached a sink"
}
