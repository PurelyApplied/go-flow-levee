// Package receivers contains test-cases for testing PII leak detection when sources are introduced via receivers.
package receivers

import (
	"example.com/core"
)

type clusterBuilder struct {
	clusterP *core.Source
	clusterV core.Source
}

func (b *clusterBuilder) buildP() {
	core.Sinkf("Building cluster %v", b.clusterP) // want "a source has reached a sink"
	core.Sinkf("Building cluster %v", b.clusterV) // want "a source has reached a sink"
}

func (b clusterBuilder) buildV() {
	core.Sinkf("Building cluster %v", b.clusterP) // want "a source has reached a sink"
	core.Sinkf("Building cluster %v", b.clusterV) // want "a source has reached a sink"
}
