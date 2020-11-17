package config

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-flow-levee/internal/pkg/config/regexp"
)

type stringMatcher interface {
	MatchString(string) bool
}

type literalMatcher string

func (lm literalMatcher) MatchString(s string) bool {
	return string(lm) == s
}

type vacuousMatcher struct{}

func (vacuousMatcher) MatchString(string) bool {
	return true
}

// Returns the first non-nil matcher.  If all are nil, returns a vacuousMatcher.
func matcherFrom(lm *literalMatcher, r *regexp.Regexp) stringMatcher {
	switch {
	case lm != nil:
		return lm
	case r != nil:
		return r
	default:
		return vacuousMatcher{}
	}
}

// A sourceMatcher matches by package, type, and field.
// Matching may be done against string literals Package, Type, Field,
// or against regexp PackageRE, TypeRE, FieldRE.
type sourceMatcher struct {
	Package stringMatcher
	Type    stringMatcher
	Field   stringMatcher
	Exclude []*sourceMatcher
}

// this type uses the default unmarshaler and mirrors configuration key-value pairs
type rawSourceMatcher struct {
	Package   *literalMatcher
	Type      *literalMatcher
	Field     *literalMatcher
	PackageRE *regexp.Regexp
	TypeRE    *regexp.Regexp
	FieldRE   *regexp.Regexp
	Exclude   []*sourceMatcher
}

func (s *sourceMatcher) UnmarshalJSON(bytes []byte) error {
	raw := rawSourceMatcher{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}

	// validation: do not double-specify any attribute with literal and regexp
	if raw.Package != nil && raw.PackageRE != nil {
		return fmt.Errorf("expected only one of Package, PackageRE to be configured")
	}
	if raw.Type != nil && raw.TypeRE != nil {
		return fmt.Errorf("expected only one of Type, TypeRE to be configured")
	}
	if raw.Field != nil && raw.FieldRE != nil {
		return fmt.Errorf("expected only one of Field, FieldRE to be configured")
	}

	*s = sourceMatcher{
		Package: matcherFrom(raw.Package, raw.PackageRE),
		Type:    matcherFrom(raw.Type, raw.TypeRE),
		Field:   matcherFrom(raw.Field, raw.FieldRE),
		Exclude: raw.Exclude,
	}
	return nil
}

func (s sourceMatcher) MatchType(path, typeName string) bool {
	if match := s.Package.MatchString(path) && s.Type.MatchString(typeName); !match {
		return false
	}

	for _, ex := range s.Exclude {
		if ex.MatchType(path, typeName) {
			return false
		}
	}

	return true
}

func (s sourceMatcher) MatchField(path, typeName, fieldName string) bool {
	if match := s.Package.MatchString(path) &&
		s.Type.MatchString(typeName) &&
		s.Field.MatchString(fieldName); !match {
		return false
	}

	for _, ex := range s.Exclude {
		if ex.MatchField(path, typeName, fieldName) {
			return false
		}
	}
	return true
}

type funcMatcher struct {
	Package  stringMatcher
	Receiver stringMatcher
	Method   stringMatcher
	Exclude  []*funcMatcher
}

// this type uses the default unmarshaler and mirrors configuration key-value pairs
type rawFuncMatcher struct {
	Package    *literalMatcher
	Receiver   *literalMatcher
	Method     *literalMatcher
	PackageRE  *regexp.Regexp
	ReceiverRE *regexp.Regexp
	MethodRE   *regexp.Regexp
	Exclude    []*funcMatcher
}

func (fm *funcMatcher) UnmarshalJSON(bytes []byte) error {
	raw := rawFuncMatcher{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}

	// validation: do not double-specify any attribute with literal and regexp
	if raw.Package != nil && raw.PackageRE != nil {
		return fmt.Errorf("expected at most one of Package, PackageRE to be configured")
	}
	if raw.Receiver != nil && raw.ReceiverRE != nil {
		return fmt.Errorf("expected at most one of Receiver, ReceiverRE to be configured")
	}
	if raw.Method != nil && raw.MethodRE != nil {
		return fmt.Errorf("expected at most one of Method, MethodRE to be configured")
	}

	*fm = funcMatcher{
		Package:  matcherFrom(raw.Package, raw.PackageRE),
		Receiver: matcherFrom(raw.Receiver, raw.ReceiverRE),
		Method:   matcherFrom(raw.Method, raw.MethodRE),
		Exclude:  raw.Exclude,
	}
	return nil
}

func (fm funcMatcher) MatchFunction(path, receiver, name string) bool {
	if match := fm.Package.MatchString(path) &&
		fm.Receiver.MatchString(receiver) &&
		fm.Method.MatchString(name); !match {
		return false
	}

	for _, ex := range fm.Exclude {
		if ex.MatchFunction(path, receiver, name) {
			return false
		}
	}

	return true
}
