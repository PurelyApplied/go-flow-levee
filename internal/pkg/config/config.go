// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"sigs.k8s.io/yaml"

	"github.com/google/go-flow-levee/internal/pkg/config/regexp"
)

// FlagSet should be used by analyzers to reuse -config flag.
var FlagSet flag.FlagSet
var configFile string

func init() {
	FlagSet.StringVar(&configFile, "config", "config.yaml", "path to analysis configuration file")
}

// config contains matchers and analysis scope information
type Config struct {
	Sources    []sourceMatcher
	Sinks      []funcMatcher
	Sanitizers []funcMatcher
	FieldTags  []fieldTagMatcher
	Exclude    []funcMatcher
}

type fieldTagMatcher struct {
	Key string
	Val string
}

// IsSourceFieldTag determines whether a field tag made up of a key and value
// is a Source.
func (c Config) IsSourceFieldTag(tag string) bool {
	if unq, err := strconv.Unquote(tag); err == nil {
		tag = unq
	}
	st := reflect.StructTag(tag)

	// built in
	if st.Get("levee") == "source" {
		return true
	}
	// configured
	for _, ft := range c.FieldTags {
		val := st.Get(ft.Key)
		for _, v := range strings.Split(val, ",") {
			if v == ft.Val {
				return true
			}
		}
	}
	return false
}

// IsExcluded determines if a function matches one of the exclusion patterns.
func (c Config) IsExcluded(path, recv, name string) bool {
	for _, pm := range c.Exclude {
		if pm.MatchFunction(path, recv, name) {
			return true
		}
	}
	return false
}

func (c Config) IsSink(path, recv, name string) bool {
	for _, p := range c.Sinks {
		if p.MatchFunction(path, recv, name) {
			return true
		}
	}
	return false
}

func (c Config) IsSanitizer(path, recv, name string) bool {
	for _, p := range c.Sanitizers {
		if p.MatchFunction(path, recv, name) {
			return true
		}
	}
	return false
}

func (c Config) IsSourceType(path, name string) bool {
	for _, p := range c.Sources {
		if p.MatchType(path, name) {
			return true
		}
	}
	return false
}

func (c Config) IsSourceField(path, typeName, fieldName string) bool {
	for _, p := range c.Sources {
		if p.MatchField(path, typeName, fieldName) {
			return true
		}
	}
	return false
}

// A sourceMatcher matches by package, type, and field.
// Matching may be done against string literals Package, Type, Field,
// or against regexp PackageRE, TypeRE, FieldRE.
type sourceMatcher struct {
	Package   *string
	Type      *string
	Field     *string
	PackageRE *regexp.Regexp
	TypeRE    *regexp.Regexp
	FieldRE   *regexp.Regexp
}

// this type uses the default unmarshaler
type rawSourceMatcher sourceMatcher

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

	// Copy all fields from raw
	*s = sourceMatcher(raw)
	return nil
}

func (s sourceMatcher) MatchType(path, typeName string) bool {
	return matchEither(s.Package, s.PackageRE, path) && matchEither(s.Type, s.TypeRE, typeName)
}

func (s sourceMatcher) MatchField(path, typeName, fieldName string) bool {
	return s.MatchType(path, typeName) && matchEither(s.Field, s.FieldRE, fieldName)
}

type funcMatcher struct {
	Package    *string
	Receiver   *string
	Method     *string
	PackageRE  *regexp.Regexp
	ReceiverRE *regexp.Regexp
	MethodRE   *regexp.Regexp
}

type rawFuncMatcher funcMatcher

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

	// copy all fields from raw
	*fm = funcMatcher(raw)
	return nil
}

func (fm funcMatcher) MatchFunction(path, receiver, name string) bool {
	return matchEither(fm.Package, fm.PackageRE, path) &&
		matchEither(fm.Receiver, fm.ReceiverRE, receiver) &&
		matchEither(fm.Method, fm.MethodRE, name)
}

// TODO This is a terrible name.  matchAnyOrNil is not better.
// Matches match against a string literal or regexp.
// Returns vacuous true when both matchers are nil.
func matchEither(literal *string, r *regexp.Regexp, match string) bool {
	return literal == nil && r == nil || literal != nil && *literal == match || r != nil && r.MatchString(match)
}

var readFileOnce sync.Once
var readConfigCached *Config
var readConfigCachedErr error

func ReadConfig() (*Config, error) {
	readFileOnce.Do(func() {
		c := new(Config)
		bytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			readConfigCachedErr = fmt.Errorf("error reading analysis config: %v", err)
			return
		}

		if err := yaml.UnmarshalStrict(bytes, c); err != nil {
			readConfigCachedErr = err
			return
		}
		readConfigCached = c
	})
	return readConfigCached, readConfigCachedErr
}
