// Package fields contains tests related to fields accessors
package fields

import (
	"google3/base/go/log"
	"google3/cloud/kubernetes/engine/common/sanitize"

	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
	hmspb "google3/google/container/v1/hosted_master_service_go_proto"
)

func testProtoStyleFieldAccessorPII(c *ipb.Cluster) {
	log.Infof("MasterAuth: %v", c.GetMasterAuth()) // want "a source has reached a sink"
}

func testDirectFieldAccessorPII(c *ipb.Cluster, trs *hmspb.TokenReviewSpec,
	authReq *hmspb.AuthenticationRequest, authResp *hmspb.AuthenticationResponse) {
	log.Infof("MasterAuth: %v", c.MasterAuth)              // want "a source has reached a sink"
	log.Infof("TokenReviewSpec: %v", trs.Token)            // want "a source has reached a sink"
	log.Infof("AuthenticationRequest: %v", authReq.Spec)   // want "a source has reached a sink"
	log.Infof("AuthenticationResponse: %v", authResp.Spec) // want "a source has reached a sink"
}

func testProtoStyleFieldAccessorSanitizedPII(c *ipb.Cluster) {
	log.Infof("MasterAuth: %v", sanitize.MakeSafeForLogs(c.GetMasterAuth()))
}

func testProtoStyleFieldAccessorNonPII(c *ipb.Cluster) {
	log.Infof("ClusterHash: %v", c.GetClusterHash())
}

func testProtoStyleFieldAccessorPIISecondLevel(c *ipb.Cluster) {
	log.Infof("MasterAuth Password: %v", c.MasterAuth.GetPassword())      // want "a source has reached a sink"
	log.Infof("MasterAuth Password: %v", c.GetMasterAuth().GetPassword()) // want "a source has reached a sink"
}

func tesDirectFieldAccessorPIISecondLevel(c *ipb.Cluster, trs *hmspb.TokenReviewSpec) {
	log.Infof("MasterAuth.Password: %v", c.MasterAuth.Password)             // want "a source has reached a sink"
	log.Infof("MasterAuth.Token: %v", c.MasterAuth.Token)                   // want "a source has reached a sink"
	log.Infof("MasterAuth.KubeProxyToken: %v", c.MasterAuth.KubeProxyToken) // want "a source has reached a sink"
	log.Infof("MasterAuth.OldToken: %v", c.MasterAuth.OldToken)             // want "a source has reached a sink"
	log.Infof("MasterAuth.ClientKey: %v", c.MasterAuth.ClientKey)           // want "a source has reached a sink"
	log.Infof("MasterAuth.MasterKey: %v", c.MasterAuth.MasterKey)           // want "a source has reached a sink"
	log.Infof("MasterAuth.KubeletKey: %v", c.MasterAuth.KubeletKey)         // want "a source has reached a sink"

	log.Infof("MasterAuth.Password: %v", c.GetMasterAuth().Password)             // want "a source has reached a sink"
	log.Infof("MasterAuth.Token: %v", c.GetMasterAuth().Token)                   // want "a source has reached a sink"
	log.Infof("MasterAuth.KubeProxyToken: %v", c.GetMasterAuth().KubeProxyToken) // want "a source has reached a sink"
	log.Infof("MasterAuth.OldToken: %v", c.GetMasterAuth().OldToken)             // want "a source has reached a sink"
	log.Infof("MasterAuth.ClientKey: %v", c.GetMasterAuth().ClientKey)           // want "a source has reached a sink"
	log.Infof("MasterAuth.MasterKey: %v", c.GetMasterAuth().MasterKey)           // want "a source has reached a sink"
	log.Infof("MasterAuth.KubeletKey: %v", c.GetMasterAuth().KubeletKey)         // want "a source has reached a sink"
}

func tesDirectFieldAccessorNoPIISecondLevel(c *ipb.Cluster) {
	log.Infof("MasterAuth.User: %v", c.MasterAuth.User)
	log.Infof("MasterAuth.ClientCertificate: %v", c.MasterAuth.ClientCertificate)

	log.Infof("MasterAuth.User: %v", c.GetMasterAuth().User)
	log.Infof("MasterAuth.ClientCertificate: %v", c.GetMasterAuth().ClientCertificate)
}

func testDirectFieldAccessorNoPII(c *ipb.Cluster,
	authReq *hmspb.AuthenticationRequest, authResp *hmspb.AuthenticationResponse) {
	log.Infof("ProjectName: %v", c.ProjectName)
	log.Infof("ProjectNumber: %v", c.ProjectNumber)
	log.Infof("GetClusterHash: %v", c.ClusterHash)
	log.Infof("Zone: %v", c.Zone)
	log.Infof("Name: %v", c.Name)

	log.Infof("AuthenticationRequest.Kind: %v", authReq.Kind)
	log.Infof("AuthenticationResponse.Kind: %v", authResp.Kind)

}
