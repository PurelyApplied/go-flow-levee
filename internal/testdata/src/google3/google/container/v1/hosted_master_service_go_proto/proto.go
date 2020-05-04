// Package hosted_master_service fakes google3/google/container/v1/hosted_master_service.proto generated files
package hosted_master_service_go_proto // NOLINT: must match the production package name of generated package.

// TokenReviewSpec fakes TokenReviewSpec
type TokenReviewSpec struct {
	Token string
}

// AuthenticationRequest fakes AuthenticationRequest
type AuthenticationRequest struct {
	Spec TokenReviewSpec
	Kind string
}

// AuthenticationResponse fakes AuthenticationResponse
type AuthenticationResponse struct {
	Spec TokenReviewSpec
	Kind string
}
