// Package membership contains test-cases for testing PII leak detection when sources are contained in structs, tuples, etc
package membership

import (
	"google3/base/go/log"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

type piiContainer struct {
	nonPiiFormatStr string
	piiVal          ipb.Cluster
	piiPtr          *ipb.Cluster
}

func piiValTuple(pc piiContainer) (string, ipb.Cluster) {
	return pc.nonPiiFormatStr, pc.piiVal
}

func piiPtrTuple(pc piiContainer) (string, *ipb.Cluster) {
	return pc.nonPiiFormatStr, pc.piiPtr
}

func piiContainerValueDetected(pc piiContainer, pcPtr *piiContainer) {
	log.Infof("piiContainer: %v", pc)    // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("piiContainer: %v", pcPtr) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func tupleContainerDetected(pc piiContainer) {
	// Tuples are automatically unpacked when passed to varargs, here passing PII to print.
	log.Infof(piiValTuple(pc)) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof(piiPtrTuple(pc)) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func wrappedPIIDetected(piiArray [5]ipb.Cluster, piiSlice []ipb.Cluster, piiMapOut map[string]ipb.Cluster, piiMapIn map[ipb.Cluster]string, piiChan chan ipb.Cluster) {
	log.Infof("Array: %v", piiArray)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("Slice: %v", piiSlice)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("MapOut: %v", piiMapOut)           // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("MapIn: %v", piiMapIn)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("Chan (recieving): %v", <-piiChan) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}

func wrappedPIIContainersDetected(piiArray [5]piiContainer, piiSlice []piiContainer, piiMapOut map[string]piiContainer, piiMapIn map[piiContainer]string, piiChan chan piiContainer) {
	log.Infof("Array: %v", piiArray)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("Slice: %v", piiSlice)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("MapOut: %v", piiMapOut)           // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("MapIn: %v", piiMapIn)             // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
	log.Infof("Chan (recieving): %v", <-piiChan) // TODO(patrhom) want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}
