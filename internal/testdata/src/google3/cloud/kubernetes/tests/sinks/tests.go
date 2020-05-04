// Package sinks contains test-cases for testing PII leak configuration regexp.
package sinks

import (
	"fmt"

	"google3/base/go/log"
	"google3/util/task/go/status"

	"google3/cloud/kubernetes/engine/common/errdesc"
	ipb "google3/cloud/kubernetes/proto/internal_go_proto"
)

type fakeWriter struct{}

func (f fakeWriter) Write(p []byte) (n int, err error) {
	panic("implement me")
}

func testFmtSinks(c *ipb.Cluster) {
	_ = fmt.Errorf("error with %v", c) // want "a source has reached a sink"

	w := fakeWriter{}
	_, _ = fmt.Fprint(w, c)                 // want "a source has reached a sink"
	_, _ = fmt.Fprintf(w, "cluster: %v", c) // want "a source has reached a sink"
	_, _ = fmt.Fprintln(w, c)               // want "a source has reached a sink"

	fmt.Print(c)                 // want "a source has reached a sink"
	fmt.Printf("cluster: %v", c) // want "a source has reached a sink"
	fmt.Println(c)               // want "a source has reached a sink"

	_ = fmt.Sprint(c)
	_ = fmt.Sprintf("cluster: %v", c)
	_ = fmt.Sprintln(c)
}

func testLogSinks(c *ipb.Cluster) {
	log.Info(c)                         // want "a source has reached a sink"
	log.InfoDepth(1, c)                 // want "a source has reached a sink"
	log.InfoDepthf(1, "cluster: %v", c) // want "a source has reached a sink"
	log.Infoln(c)                       // want "a source has reached a sink"
	log.Infof("cluster: %v", c)         // want "a source has reached a sink"

	log.Warning(c)                         // want "a source has reached a sink"
	log.WarningDepth(1, c)                 // want "a source has reached a sink"
	log.WarningDepthf(1, "cluster: %v", c) // want "a source has reached a sink"
	log.Warningln(c)                       // want "a source has reached a sink"
	log.Warningf("cluster: %v", c)         // want "a source has reached a sink"

	log.Error(c)                         // want "a source has reached a sink"
	log.ErrorDepth(1, c)                 // want "a source has reached a sink"
	log.ErrorDepthf(1, "cluster: %v", c) // want "a source has reached a sink"
	log.Errorln(c)                       // want "a source has reached a sink"
	log.Errorf("cluster: %v", c)         // want "a source has reached a sink"

	log.Fatal(c)                         // want "a source has reached a sink"
	log.FatalDepth(1, c)                 // want "a source has reached a sink"
	log.FatalDepthf(1, "cluster: %v", c) // want "a source has reached a sink"
	log.Fatalln(c)                       // want "a source has reached a sink"
	log.Fatalf("cluster: %v", c)         // want "a source has reached a sink"

	log.Exit(c)                         // want "a source has reached a sink"
	log.ExitDepth(1, c)                 // want "a source has reached a sink"
	log.ExitDepthf(1, "cluster: %v", c) // want "a source has reached a sink"
	log.Exitln(c)                       // want "a source has reached a sink"
	log.Exitf("cluster: %v", c)         // want "a source has reached a sink"
}

func testErrdescSinks(c *ipb.Cluster) {
	_ = errdesc.InternalError.WithMsg("error with cluster %v", c) // want "a source has reached a sink"
}

func testStatusSinks(c *ipb.Cluster) {
	_ = status.Errorf(status.FakeCode, "error with cluster %v", c)                               // want "a source has reached a sink"
	_ = status.Errorf(status.FakeCode, "error with cluster password %v", *c.MasterAuth.Password) // want "a source has reached a sink"
	_ = status.Error(status.FakeCode, *c.MasterAuth.Password)                                    // TODO(b/148147663): Only variadic sink arguments are currently detected. -- want "input to logging methods must be sanitized via cloud/kubernetes/engine/common/sanitize package"
}
