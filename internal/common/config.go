package common

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/types"
	"io/ioutil"
	"regexp"
	"sync"

	"golang.org/x/tools/go/ssa"
)

// config contains matchers and analysis scope information
type Config struct {
	Sources                 []sourceMatcher
	Sinks                   []NameMatcher
	Sanitizers              []NameMatcher
	FieldPropagators        []fieldPropagatorMatcher
	TransformingPropagators []transformingPropagatorMatcher
	PropagatorArgs          argumentPropagatorMatcher
	Whitelist               []packageMatcher
	AnalysisScope           []packageMatcher
}

// shouldSkip returns true for any function that is outside analysis scope,
// that is whitelisted,
// whose containing package imports "testing"
// or whose containing package does not import any package containing a source or a sink.
func (c Config) shouldSkip(pkg *types.Package) bool {
	if IsTestPkg(pkg) || !c.isInScope(pkg) || c.isWhitelisted(pkg) {
		return true
	}

	// TODO Does this skip packages that own sources/sinks but don't import others?
	for _, im := range pkg.Imports() {
		for _, s := range c.Sinks {
			if s.matchPackage(im) {
				return false
			}
		}

		for _, s := range c.Sources {
			if s.PackageRE.MatchString(im.Path()) {
				return false
			}
		}
	}

	return true
}

func (c Config) IsSink(call *ssa.Call) bool {
	for _, p := range c.Sinks {
		if p.MatchMethodName(call) {
			return true
		}
	}

	return false
}

func (c Config) IsSanitizer(call *ssa.Call) bool {
	for _, p := range c.Sanitizers {
		if p.MatchMethodName(call) {
			return true
		}
	}

	return false
}

func (c Config) IsSource(t types.Type) bool {
	n, ok := t.(*types.Named)
	if !ok {
		return false
	}

	for _, p := range c.Sources {
		if p.match(n) {
			return true
		}
	}
	return false
}

func (c Config) IsSourceFieldAddr(fa *ssa.FieldAddr) bool {
	// fa.Type() refers to the accessed field's type.
	// fa.X.Type() refers to the surrounding struct's type.

	deref := DereferenceRecursive(fa.X.Type())
	st, ok := deref.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	fieldName := st.Field(fa.Field).Name()

	for _, p := range c.Sources {
		if n, ok := deref.(*types.Named); ok &&
			p.match(n) && p.FieldRE.MatchString(fieldName) {
			return true
		}
	}
	return false
}

func (c Config) IsPropagator(call *ssa.Call) bool {
	return c.isFieldPropagator(call) || c.isTransformingPropagator(call)
}

func (c Config) isFieldPropagator(call *ssa.Call) bool {
	recv := call.Call.Signature().Recv()
	if recv == nil {
		return false
	}

	for _, p := range c.FieldPropagators {
		if p.match(call) {
			return true
		}
	}

	return false
}

func (c Config) isTransformingPropagator(call *ssa.Call) bool {
	for _, p := range c.TransformingPropagators {
		if !p.match(call) {
			continue
		}

		for _, a := range call.Call.Args {
			// TODO Handle ChangeInterface case.
			switch t := a.(type) {
			case *ssa.MakeInterface:
				if c.IsSource(DereferenceRecursive(t.X.Type())) {
					return true
				}
			case *ssa.Parameter:
				if c.IsSource(DereferenceRecursive(t.Type())) {
					return true
				}
			}
		}
	}

	return false
}

func (c Config) SendsToIOWriter(call *ssa.Call) ssa.Node {
	if call.Call.Signature().Params().Len() == 0 {
		return nil
	}

	firstArg := call.Call.Signature().Params().At(0)
	if c.PropagatorArgs.ArgumentTypeRE.MatchString(firstArg.Type().String()) {
		if a, ok := call.Call.Args[0].(*ssa.MakeInterface); ok {
			return a.X.(ssa.Node)
		}
	}

	return nil
}

func (c Config) isWhitelisted(pkg *types.Package) bool {
	for _, w := range c.Whitelist {
		if w.match(pkg) {
			return true
		}
	}
	return false
}

func (c Config) isInScope(pkg *types.Package) bool {
	for _, s := range c.AnalysisScope {
		if s.match(pkg) {
			return true
		}
	}
	return false
}

// A sourceMatcher defines what types are or contain sources.
// Within a given type, specific field access can be specified as the actual source data
// via the fieldRE.
type sourceMatcher struct {
	PackageRE configRegexp
	TypeRE    configRegexp
	FieldRE   configRegexp
}

func (s sourceMatcher) match(n *types.Named) bool {
	if types.IsInterface(n) {
		// In our context, both sources and sanitizers are concrete types.
		return false
	}

	return s.PackageRE.MatchString(n.Obj().Pkg().Path()) && s.TypeRE.MatchString(n.Obj().Name())
}

type fieldPropagatorMatcher struct {
	Receiver   string
	AccessorRE configRegexp
}

func (f fieldPropagatorMatcher) match(call *ssa.Call) bool {
	if call.Call.StaticCallee() == nil {
		return false
	}

	recv := call.Call.Signature().Recv()
	if recv == nil {
		return false
	}

	if f.Receiver != DereferenceRecursive(recv.Type()).String() {
		return false
	}

	return f.AccessorRE.MatchString(call.Call.StaticCallee().Name())
}

type transformingPropagatorMatcher struct {
	PackageName string
	MethodRE    configRegexp
}

func (t transformingPropagatorMatcher) match(call *ssa.Call) bool {
	if call.Call.StaticCallee() == nil ||
		call.Call.StaticCallee().Pkg == nil ||
		call.Call.StaticCallee().Pkg.Pkg.Path() != t.PackageName {
		return false
	}

	return t.MethodRE.MatchString(call.Call.StaticCallee().Name())
}

type argumentPropagatorMatcher struct {
	ArgumentTypeRE configRegexp
}

type packageMatcher struct {
	PackageNameRE configRegexp
}

func (pm packageMatcher) match(pkg *types.Package) bool {
	return pm.PackageNameRE.MatchString(pkg.Path())
}

// configRegexp delegates to a Regexp while enabling unmarshalling.
// Any unspecified / nil matcher will return vacuous truth in MatchString
type configRegexp struct {
	r *regexp.Regexp
}

func (mr *configRegexp) MatchString(s string) bool {
	return mr.r == nil || mr.r.MatchString(s)
}

func (mr *configRegexp) UnmarshalJSON(data []byte) error {
	var matcher string
	if err := json.Unmarshal(data, &matcher); err != nil {
		return err
	}

	var err error
	if mr.r, err = regexp.Compile(matcher); err != nil {
		return err
	}
	return nil
}

type NameMatcher struct {
	PackageRE configRegexp
	TypeRE    configRegexp
	MethodRE  configRegexp
}

func (r NameMatcher) matchPackage(p *types.Package) bool {
	return r.PackageRE.MatchString(p.Path())
}

func (r NameMatcher) MatchMethodName(c *ssa.Call) bool {
	if c.Call.StaticCallee() == nil || c.Call.StaticCallee().Pkg == nil {
		return false
	}

	return r.matchPackage(c.Call.StaticCallee().Pkg.Pkg) &&
		r.MethodRE.MatchString(c.Call.StaticCallee().Name())
}

func (r NameMatcher) matchNamedType(n *types.Named) bool {
	if types.IsInterface(n) {
		// In our context, both sources and sanitizers are concrete types.
		return false
	}

	return r.PackageRE.MatchString(n.Obj().Pkg().Path()) &&
		r.TypeRE.MatchString(n.Obj().Name())
}

var readFileOnce sync.Once
var readConfigCached *Config
var readConfigCachedErr error

var SharedFlags flag.FlagSet
var configFile string

func init() {
	SharedFlags.StringVar(&configFile, "config", "config.json", "path to taint propagation analysis configuration file")
}

func readConfig(path string) (*Config, error) {
	loadedFromCache := true
	readFileOnce.Do(func() {
		loadedFromCache = false
		c := new(Config)
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			readConfigCachedErr = fmt.Errorf("error reading analysis config: %v", err)
		}

		if err := json.Unmarshal(bytes, c); err != nil {
			readConfigCachedErr = err
		}
		readConfigCached = c
	})
	_ = loadedFromCache
	return readConfigCached, readConfigCachedErr
}
