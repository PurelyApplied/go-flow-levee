// Package fields contains tests related to fields accessors
package fields

import (
	"example.com/core"
)

func TestFieldAccessors(s core.Source, ptr *core.Source) {
	core.Sinkf("Data: %v", s.GetData()) // want "a source has reached a sink"
	core.Sinkf("ID: %v", s.GetID())

	core.Sinkf("Data: %v", ptr.GetData()) // want "a source has reached a sink"
	core.Sinkf("ID: %v", ptr.GetID())
}

func TestDirectFieldAccess(c *core.Source) {
	core.Sinkf("Data: %v", c.Data) // want "a source has reached a sink"
	core.Sinkf("ID: %v", c.ID)
}

func testProtoStyleFieldAccessorSanitizedPII(c *core.Source) {
	core.Sinkf("MasterAuth: %v", core.Sanitize(c.GetData()))
}

func testProtoStyleFieldAccessorPIISecondLevel(wrapper struct{ *core.Source }) {
	core.Sinkf("MasterAuth Password: %v", wrapper.Source.GetData()) // want "a source has reached a sink"
	core.Sinkf("MasterAuth Password: %v", wrapper.Source.GetID())
}

func tesDirectFieldAccessorPIISecondLevel(wrapper struct{ *core.Source }) {
	core.Sinkf("MasterAuth.Password: %v", wrapper.Source.Data) // want "a source has reached a sink"
	core.Sinkf("MasterAuth.Password: %v", wrapper.Source.ID)
}
