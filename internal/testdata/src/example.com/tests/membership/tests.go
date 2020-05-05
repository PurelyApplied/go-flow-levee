// Package membership contains test-cases for testing PII leak detection when sources are contained in structs, tuples, etc
package membership

import (
	"example.com/core"
)

type sourceContainer struct {
	fmtString string
	sourceVal core.Source
	sourcePtr *core.Source
}

func piiValTuple(pc sourceContainer) (string, core.Source) {
	return pc.fmtString, pc.sourceVal
}

func piiPtrTuple(pc sourceContainer) (string, *core.Source) {
	return pc.fmtString, pc.sourcePtr
}

func piiContainerValueDetected(pc sourceContainer, pcPtr *sourceContainer) {
	core.Sinkf("sourceContainer: %v", pc)    // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("sourceContainer: %v", pcPtr) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func tupleContainerDetected(pc sourceContainer) {
	// Tuples are automatically unpacked when passed to varargs, here passing PII to print.
	core.Sinkf(piiValTuple(pc)) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf(piiPtrTuple(pc)) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func wrappedPIIDetected(piiArray [5]core.Source, piiSlice []core.Source, piiMapOut map[string]core.Source, piiMapIn map[core.Source]string, piiChan chan core.Source) {
	core.Sinkf("Array: %v", piiArray)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("Slice: %v", piiSlice)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("MapOut: %v", piiMapOut)           // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("MapIn: %v", piiMapIn)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("Chan (recieving): %v", <-piiChan) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func wrappedPIIContainersDetected(piiArray [5]sourceContainer, piiSlice []sourceContainer, piiMapOut map[string]sourceContainer, piiMapIn map[sourceContainer]string, piiChan chan sourceContainer) {
	core.Sinkf("Array: %v", piiArray)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("Slice: %v", piiSlice)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("MapOut: %v", piiMapOut)           // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("MapIn: %v", piiMapIn)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("Chan (recieving): %v", <-piiChan) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}
