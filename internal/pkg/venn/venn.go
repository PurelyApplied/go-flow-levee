package venn

import (
	"fmt"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var hackyFilter = "ssa"

var Analyzer = &analysis.Analyzer{
	Name: "venn",
	Doc:  "Builds a Venn Diagram of a package's types and interfaces",
	Run:  run,
}

type implKind uint
const(
	no implKind = 0
	value implKind = iota
	reference
)

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != hackyFilter {
		return nil, nil
	}

	scope := pass.Pkg.Scope()
	names := scope.Names()
	namedStructs := make(map[*types.Named]string)
	namedIface := make(map[*types.Named]string)

	for _, name := range names {
		obj := scope.Lookup(name)
		if tn, ok := obj.(*types.TypeName); ok && obj.Exported() {
			typ := tn.Type()
			named, ok := typ.(*types.Named)
			if ok {
				under := named.Underlying()
				switch under.(type) {
				case *types.Struct:
					namedStructs[named] = named.String()
				case *types.Interface:
					namedIface[named] = named.String()
				default:
					fmt.Println("Got a weird named type:", named)
				}
			}
		}
	}

	printCSV(namedStructs, namedIface)
	printListing(namedStructs, namedIface)
	//return nil, nil
}

func printCSV(namedStructs, namedIface map[*types.Named]string) {
	table := make(map[string]map[string]implKind)
	for st, stname := range namedStructs {
		table[stname] = make(map[string]implKind)
		for iface, ifname := range namedIface {
			switch {
			case types.Implements(st, iface.Underlying().(*types.Interface)):
				table[stname][ifname] = value
			case types.Implements(types.NewPointer(st), iface.Underlying().(*types.Interface)):
				table[stname][ifname] = reference
			default:
				table[stname][ifname] = no
			}
		}
	}

	strNames := sortedStr(table)
	ifNames := sortedIfa(table)
	var builder strings.Builder

	builder.WriteString(`# <type> \ <interface>:,`)
	for _, i := range ifNames {
		builder.WriteString(i+",")
	}
	builder.WriteString("\n")

	for _, s := range strNames {
		builder.WriteString(s+",")
		for _, i := range ifNames {
			switch table[s][i] {
			case no:
			case reference:
				builder.WriteString("ptr")
			case value:
				builder.WriteString("val")
			}
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}
	fmt.Println(builder.String())
}

func sortedStr(kinds map[string]map[string]implKind) []string {
	var keys []string
	for k := range kinds {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedIfa(kinds map[string]map[string]implKind) []string {
	var inames []string
	for _, is := range kinds {
		for iface := range is {
			inames = append(inames, iface)
		}
		break
	}

	sort.Strings(inames)
	return inames
}

func printListing(namedStructs, namedIface map[*types.Named]string) {
	impls := make(map[string][]string)
	implBy := make(map[string][]string)
	for _, name := range namedStructs {
		impls[name] = []string{}
	}
	for _, name := range namedIface {
		implBy[name] = []string{}
	}

	for st, stname := range namedStructs {
		for iface, iname := range namedIface {
			under := iface.Underlying().(*types.Interface)
			implements := types.Implements(st, under)
			if implements {
				impls[stname] = append(impls[stname], iname)
				implBy[iname] = append(implBy[iname], stname)
			} else {

				pointer := types.NewPointer(st)
				ptrImpl := types.Implements(pointer, under)
				if ptrImpl {
					impls[stname] = append(impls[stname], "*" + iname)
					implBy[iname] = append(implBy[iname], "*" + stname)
				}
			}
		}
	}

	fmt.Println("The following named structs implement the following named interfaces:")
	for _, st := range keys(impls) {
		imps := impls[st]
		sort.Strings(imps)
		fmt.Println("  ", st)
		for _, iface := range imps {
			fmt.Println("      ", iface)
		}
	}

	fmt.Println("The following named interfaces are implemented by the following named structs:")
	for _, ifac := range keys(implBy) {
		strucs := implBy[ifac]
		sort.Strings(strucs)
		fmt.Println("  ", ifac)
		for _, s := range strucs {
			fmt.Println("      ", s)
		}
	}


}

func keys(m map[string][]string)[]string{
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}