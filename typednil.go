package typednil

import (
	"fmt"
	"go/token"
	"go/types"

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
	FactTypes: []analysis.Fact{new(isTypedNilFunc)},
}

type isTypedNilFunc struct {
	Index []int
}

func (*isTypedNilFunc) AFact() {}

var _ analysis.Fact = (*isTypedNilFunc)(nil)

func (f *isTypedNilFunc) String() string {
	return fmt.Sprintf("isTypedFunc%v", f.Index)
}

var _ fmt.Stringer = (*isTypedNilFunc)(nil)

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
		cnst, _ := v.X.(*ssa.Const)
		return cnst != nil && cnst.IsNil() && !types.Identical(cnst.Type(), types.Typ[types.UntypedNil])
	case *ssa.Call:
		fact := a.typedNilFunc(v)
		return fact != nil && len(fact.Index) == 1
	case *ssa.Extract:
		call, _ := v.Tuple.(*ssa.Call)
		if call == nil {
			return false
		}
		fact := a.typedNilFunc(call)
		if fact == nil {
			return false
		}
		for _, i := range fact.Index {
			if i == v.Index {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (a *analyzer) typedNilFunc(v *ssa.Call) *isTypedNilFunc {
	if v.Call.Method != nil {
		return nil
	}

	fun, _ := v.Call.Value.(*ssa.Function)
	if fun == nil {
		return nil
	}

	var fact isTypedNilFunc
	ok := a.pass.ImportObjectFact(fun.Object(), &fact)
	if ok {
		return &fact
	}

	return nil
}

func (a *analyzer) isUntypedNil(v ssa.Value) bool {
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
		var index []int
		exclude := make(map[int]bool)
		for _, ret := range rets {
			for i, r := range ret.Results {
				if exclude[i] {
					continue
				}

				if !types.IsInterface(r.Type().Underlying()) {
					exclude[i] = true
					continue
				}

				if a.isTypedNil(r) {
					index = append(index, i)
					exclude[i] = true
				}
			}
		}

		if index != nil {
			fact := &isTypedNilFunc{Index: index}
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

				if (a.isTypedNil(binOp.X) && a.isUntypedNil(binOp.Y)) ||
					(a.isTypedNil(binOp.Y) && a.isUntypedNil(binOp.X)) {
					a.pass.Reportf(binOp.Pos(), "it may become a comparition a typed nil and an untyped nil")
				}
			}
		}
	}
}
