// Package membership contains test-cases for testing PII leak detection when sources are contained in structs, tuples, etc
package membership

import (
	"example.com/core"
)

type piiContainer struct {
	nonPiiFormatStr string
	piiVal          core.Source
	piiPtr          *core.Source
}

func piiValTuple(pc piiContainer) (string, core.Source) {
	return pc.nonPiiFormatStr, pc.piiVal
}

func piiPtrTuple(pc piiContainer) (string, *core.Source) {
	return pc.nonPiiFormatStr, pc.piiPtr
}

func piiContainerValueDetected(pc piiContainer, pcPtr *piiContainer) {
	core.Sinkf("piiContainer: %v", pc)    // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("piiContainer: %v", pcPtr) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func tupleContainerDetected(pc piiContainer) {
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

func wrappedPIIContainersDetected(piiArray [5]piiContainer, piiSlice []piiContainer, piiMapOut map[string]piiContainer, piiMapIn map[piiContainer]string, piiChan chan piiContainer) {
	core.Sinkf("Array: %v", piiArray)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("Slice: %v", piiSlice)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("MapOut: %v", piiMapOut)           // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("MapIn: %v", piiMapIn)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	core.Sinkf("Chan (recieving): %v", <-piiChan) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}
