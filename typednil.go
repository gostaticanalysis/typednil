package typednil

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const doc = "typednil finds a comparition between typed nil and untyped nil"

// Analyzer finds a comparition between typed nil and untyped nil
var Analyzer = &analysis.Analyzer{
	Name: "typednil",
	Doc:  doc,
	Run:  new(analyzer).run,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
	FactTypes: []analysis.Fact{new(isNilable)},
}

type nilableKind int

const (
	interfaceNilable nilableKind = iota
	concreteNilable
)

func (k nilableKind) String() string {
	switch k {
	case interfaceNilable:
		return "I"
	default:
		return "C"
	}
}

type isNilable struct {
	Results map[int]nilableKind
}

func (*isNilable) AFact() {}

var _ analysis.Fact = (*isNilable)(nil)

func (f *isNilable) String() string {
	rets := make([]string, 0, len(f.Results))
	for index, kind := range f.Results {
		rets = append(rets, fmt.Sprintf("%v:%v", index, kind))
	}
	return fmt.Sprintf("nilable results [%v]", strings.Join(rets, ","))
}

var _ fmt.Stringer = (*isNilable)(nil)

type analyzer struct {
	pass *analysis.Pass
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	a.pass = pass
	a.findTypedNilFunc()
	a.findTypedNilCmp()
	return nil, nil
}

func (a *analyzer) isTypedNil(v ssa.Value) bool {
	switch v := v.(type) {
	case *ssa.MakeInterface:
		switch x := v.X.(type) {
		case *ssa.Const:
			return x.IsNil()
		default:
			return a.nilableFuncCall(x, concreteNilable)
		}
	default:
		return a.nilableFuncCall(v, interfaceNilable)
	}
}

func (a *analyzer) nilableFuncCall(v ssa.Value, kind nilableKind) bool {
	switch v := v.(type) {
	case *ssa.Call:
		fact := a.importFact(v)
		if fact == nil {
			return false
		}
		if k, ok := fact.Results[0]; ok && k == kind {
			return true
		}
		return false
	case *ssa.Extract:
		call, _ := v.Tuple.(*ssa.Call)
		if call == nil {
			return false
		}
		fact := a.importFact(call)
		if fact == nil {
			return false
		}
		if k, ok := fact.Results[v.Index]; ok && k == kind {
			return true
		}
		return false
	default:
		return false
	}
}

func (a *analyzer) importFact(v *ssa.Call) *isNilable {
	if v.Call.Method != nil {
		return nil
	}

	fun, _ := v.Call.Value.(*ssa.Function)
	if fun == nil || fun.Object() == nil {
		return nil
	}

	var fact isNilable
	ok := a.pass.ImportObjectFact(fun.Object(), &fact)
	if ok {
		return &fact
	}

	return nil
}

func (a *analyzer) isCostNil(v ssa.Value) bool {
	switch v := v.(type) {
	case *ssa.Const:
		return v.IsNil()
	}
	return false
}

func (a *analyzer) findTypedNilFunc() {
	s := a.pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	for _, f := range s.SrcFuncs {
		obj := f.Object()
		if obj == nil {
			continue
		}

		rets := analysisutil.Returns(f)
		results := make(map[int]nilableKind)
		for _, ret := range rets {
			for i, r := range ret.Results {
				if _, ok := results[i]; ok {
					continue
				}

				switch {
				case types.IsInterface(r.Type()) && a.isTypedNil(r):
					results[i] = interfaceNilable
				case !types.IsInterface(r.Type()) && a.isCostNil(r):
					results[i] = concreteNilable
				}
			}
		}

		if len(results) != 0 {
			fact := &isNilable{Results: results}
			a.pass.ExportObjectFact(obj, fact)
		}
	}
}

func (a *analyzer) findTypedNilCmp() {
	s := a.pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	for _, f := range s.SrcFuncs {
		for _, b := range f.Blocks {
			for _, instr := range b.Instrs {
				binOp, _ := instr.(*ssa.BinOp)
				if binOp == nil ||
					(binOp.Op != token.EQL && binOp.Op != token.NEQ) {
					continue
				}

				if (a.isTypedNil(binOp.X) && a.isCostNil(binOp.Y)) ||
					(a.isTypedNil(binOp.Y) && a.isCostNil(binOp.X)) {
					a.pass.Reportf(binOp.Pos(), "it may become a comparition a typed nil and an untyped nil")
				}
			}
		}
	}
}
