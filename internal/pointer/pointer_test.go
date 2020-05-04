// This has copied callgraph/main.go wholesale while I figure some things out.
package pointer

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/tools/go/analysis/analysistest"
	"k8s.io/api/authentication/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var pkgs = []string{
	"example.com",     // package main
	"example.com/foo", // imported by main
}

func TestPointsTo(t *testing.T) {
	for _, dir := range []string{
		analysistest.TestData() + "/src",
	} {
		for _, prefix := range []string{
			"",
		} {
			t.Logf("Attempting dir %q, prefix %q", dir, prefix)
			var prefixedPkgs []string
			for _, p := range pkgs {
				prefixedPkgs = append(prefixedPkgs, prefix+p)
			}
			if _, err := getConfig(prefixedPkgs, dir, dir); err != nil {
				t.Error(err)
			}
		}
	}
}

func TestAnalyzer(t *testing.T) {
	basicConf = Config{
		Args: pkgs,
		Dir:  analysistest.TestData(),
	}
	analysistest.Run(t, analysistest.TestData(), Analyzer, pkgs...)
}

func TestFoo(t *testing.T) {
  tk := v1.TokenRequest{
		TypeMeta: v12.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:            "",
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: v12.Time{
				Time: time.Time{},
			},
			DeletionTimestamp: &v12.Time{
				Time: time.Time{},
			},
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ClusterName:                "",
			ManagedFields:              nil,
		},
		Spec: v1.TokenRequestSpec{
			Audiences:         nil,
			ExpirationSeconds: nil,
			BoundObjectRef: &v1.BoundObjectReference{
				Kind:       "",
				APIVersion: "",
				Name:       "",
				UID:        "",
			},
		},
		Status: v1.TokenRequestStatus{
			Token: "",
			ExpirationTimestamp: v12.Time{
				Time: time.Time{},
			},
		},
	}

  fmt.Printf("%#v", tk)
}