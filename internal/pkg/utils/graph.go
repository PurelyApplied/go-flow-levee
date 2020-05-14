package utils

import "golang.org/x/tools/go/ssa"

type link struct {
	from, to ssa.Node
	kind     string
}

func GetEdges(fn *ssa.Function) map[link]bool {
	edges := make(map[link]bool)
	for _, blk := range fn.Blocks {
		for _, instr := range blk.Instrs {
			for _, op := range getOps(instr) {
				if *op == nil {
					continue
				}
				l := link{
					from: instr.(ssa.Node),
					to:   (*op).(ssa.Node),
					kind: "operand",
				}
				edges[l] = true
			}

			referrers := instr.(ssa.Node).Referrers()
			if referrers != nil {
				for _, ref := range *referrers {
					l := link{
						from: instr.(ssa.Node),
						to:   ref.(ssa.Node),
						kind: "reference",
					}
					edges[l] = true
				}
			}
		}
	}
	return edges
}

func GetLinks(fn *ssa.Function) (refs, ops map[ssa.Node][]ssa.Node) {
	refGraph := make(map[ssa.Node][]ssa.Node)
	opsGraph := make(map[ssa.Node][]ssa.Node)

	edges := GetEdges(fn)
	for e := range edges {
		switch e.kind {
		case "operand":
			opsGraph[e.from] = append(opsGraph[e.from], e.to)
		case "reference":
			refGraph[e.from] = append(refGraph[e.from], e.to)
		default:
			panic("missed")
		}
	}

	return refGraph, opsGraph
}

func getOps(instr ssa.Instruction) []*ssa.Value {
	var buf [10]*ssa.Value
	ops := instr.Operands(buf[:0])
	return ops
}

func OperandsInverse(n ssa.Node, opsG map[ssa.Node][]ssa.Node) []ssa.Node {
	var x []ssa.Node
	for inv, ops := range opsG {
		for _, o := range ops {
			if o == n.(ssa.Node) {
				x = append(x, inv)
			}
		}
	}
	return x
}

func OperandsOfReferrers(n ssa.Node, refG, opsG map[ssa.Node][]ssa.Node) []ssa.Node {
	set := make(map[ssa.Node]bool)
	for _, r := range refG[n] {
		for _, o := range opsG[r] {
			set[o] = true
		}
	}

	var oor []ssa.Node
	for o := range set {
		oor = append(oor, o)
	}
	return oor
}
