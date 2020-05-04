// Package internal_go_proto fakes cloud/kubernetes/proto/internal_go_proto package.
package internal_go_proto // NOLINT: must match the production package name of generated package.

import "strconv"

// MasterAuth fakes MasterAuth
type MasterAuth struct {
	// Non-sensitive fields
	User              *string
	ClientCertificate *string

	// Sensitive fields
	Password       *string
	Token          *string
	KubeProxyToken *string
	OldToken       *string
	ClientKey      *string
	MasterKey      *string
	KubeletKey     *string
}

// ProtoMessage fakes ProtoMessage.
func (c *MasterAuth) ProtoMessage() {
}

// Reset fakes Reset.
func (c *MasterAuth) Reset() {
}

// String fakes String.
func (c *MasterAuth) String() string {
	return *c.Password
}

// GetPassword fakes GetPassword.
func (c *MasterAuth) GetPassword() *string {
	return c.Password
}

// GetToken fakes GetToken.
func (c *MasterAuth) GetToken() *string {
	return c.Token
}

// GetUser fakes GetUser
func (c *MasterAuth) GetUser() *string {
	return c.User
}

// Cluster fakes Cluster.
type Cluster struct {
	// Non-sensitive fields
	ClusterHash   string
	Zone          string
	Name          string
	ProjectName   string
	ProjectNumber int

	// Sensitive fields
	MasterAuth *MasterAuth
}

// GetClusterHash fakes GetClusterHash.
func (c *Cluster) GetClusterHash() string {
	return c.ClusterHash
}

// GetMasterAuth fakes GetMasterAuth.
func (c *Cluster) GetMasterAuth() *MasterAuth {
	return c.MasterAuth
}

// ProtoMessage fakes ProtoMessage.
func (c *Cluster) ProtoMessage() {
}

// Reset fakes Reset.
func (c *Cluster) Reset() {
}

// String fakes String.
func (c *Cluster) String() string {
	return *c.MasterAuth.Password
}

// ClusterHealth fakes non-PII type ClusterHealth
type ClusterHealth struct {
	Status int
}

// ProtoMessage fakes ProtoMessage.
func (c *ClusterHealth) ProtoMessage() {
}

// Reset fakes Reset.
func (c *ClusterHealth) Reset() {
}

// String fakes String.
func (c *ClusterHealth) String() string {
	return strconv.Itoa(c.Status)
}
