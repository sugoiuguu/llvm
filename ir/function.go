package ir

import (
	"fmt"
	"strconv"

	"github.com/llir/l/internal/enc"
	"github.com/llir/l/ir/enum"
	"github.com/llir/l/ir/types"
	"github.com/llir/l/ir/value"
	"github.com/pkg/errors"
)

// === [ Functions ] ===========================================================

// Function is an LLVM IR function.
type Function struct {
	// Function signature.
	Sig *types.FuncType
	// Function name (without '@' prefix).
	GlobalName string
	// Function parameters.
	Params []*Param
	// Basic blocks.
	Blocks []*BasicBlock

	// extra.

	// Pointer type to function, including an optional address space. If Typ is
	// nil, the first invocation of Type stores a pointer type with Sig as
	// element.
	Typ *types.PointerType
	// (optional) Linkage.
	Linkage enum.Linkage
	// (optional) Preemption; zero value if not present.
	Preemption enum.Preemption
	// (optional) Visibility; zero value if not present.
	Visibility enum.Visibility
	// (optional) DLL storage class; zero value if not present.
	DLLStorageClass enum.DLLStorageClass
	// (optional) Calling convention; zero value if not present.
	CallingConv enum.CallingConv
	// (optional) Return attributes.
	ReturnAttrs []enum.ReturnAttribute
	// (optional) Unnamed address.
	UnnamedAddr enum.UnnamedAddr
	// (optional) Function attributes.
	FuncAttrs []enum.FuncAttribute
	// (optional) Section; nil if not present.
	Section string
	// (optional) Comdat; nil if not present.
	Comdat *ComdatDef
	// (optional) Garbage collection; empty if not present.
	GC string
	// (optional) Prefix; nil if not present.
	Prefix Constant
	// (optional) Prologue; nil if not present.
	Prologue Constant
	// (optional) Personality; nil if not present.
	Personality Constant
	// (optional) Use list orders.
	// TODO: add support for UseListOrder.
	//UseListOrders []*UseListOrder
	// (optional) Metadata attachments.
	// TODO: add support for metadata.
	//Metadata []*metadata.MetadataAttachment
}

// TODO: decide whether to have the function name parameter be the first
// argument (to be consistent with NewGlobal) or after retType (to be consistent
// with order of occurence in LLVM IR syntax).

// NewFunction returns a new function based on the given function name, return
// type and function parameters.
func NewFunction(name string, retType types.Type, params ...*Param) *Function {
	panic("not yet implemented")
	/*
		paramTypes := make([]types.Type, len(params))
		for i, param := range f.Params {
			paramType[i] = param.Type()
		}
		sig := types.NewFunc(f.RetType, paramTypes...)
		return &Function{Sig: sig, GlobalName: name, Params: params}
	*/
}

// String returns the LLVM syntax representation of the function as a type-value
// pair.
func (f *Function) String() string {
	return fmt.Sprintf("%v %v", f.Type(), f.Ident())
}

// Type returns the type of the function.
func (f *Function) Type() types.Type {
	// Cache type if not present.
	if f.Typ == nil {
		f.Typ = types.NewPointer(f.Sig)
	}
	return f.Typ
}

// Ident returns the identifier associated with the function.
func (f *Function) Ident() string {
	return enc.Global(f.GlobalName)
}

// Name returns the name of the function.
func (f *Function) Name() string {
	return f.GlobalName
}

// SetName sets the name of the function.
func (f *Function) SetName(name string) {
	f.GlobalName = name
}

// Def returns the LLVM syntax representation of the function definition or
// declaration.
func (f *Function) Def() string {
	panic("not yet implemented")
}

// AssignIDs assigns IDs to unnamed local variables.
func (f *Function) AssignIDs() error {
	if len(f.Blocks) == 0 {
		return nil
	}
	id := 0
	names := make(map[string]value.Value)
	setName := func(n value.Named) error {
		got := n.Name()
		if isUnnamed(got) {
			name := strconv.Itoa(id)
			n.SetName(name)
			names[name] = n
			id++
		} else if isLocalID(got) {
			want := strconv.Itoa(id)
			if want != got {
				return errors.Errorf("invalid local ID in function %q, expected %s, got %s", enc.Global(f.GlobalName), enc.Local(want), enc.Local(got))
			}
			id++
		} else {
			// already named; nothing to do.
		}
		return nil
	}
	for _, param := range f.Params {
		// Assign local IDs to unnamed parameters of function definitions.
		if err := setName(param); err != nil {
			return errors.WithStack(err)
		}
	}
	for _, block := range f.Blocks {
		// Assign local IDs to unnamed basic blocks.
		if err := setName(block); err != nil {
			return errors.WithStack(err)
		}
		for _, inst := range block.Insts {
			n, ok := inst.(value.Named)
			if !ok {
				continue
			}
			// Skip void instructions.
			// TODO: Check if any other value instructions than call may have void
			// type.
			if isVoidValue(n) {
				continue
			}
			// Assign local IDs to unnamed local variables.
			if err := setName(n); err != nil {
				return errors.WithStack(err)
			}
		}
		n, ok := block.Term.(value.Named)
		if !ok {
			continue
		}
		if isVoidValue(n) {
			continue
		}
		if err := setName(n); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// ### [ Helper functions ] ####################################################

// isVoidValue reports whether the given named value is a non-value (i.e. a call
// instruction or invoke terminator with void-return type).
func isVoidValue(n value.Named) bool {
	switch n.(type) {
	case *InstCall, *TermInvoke:
		return n.Type().Equal(types.Void)
	}
	return false
}
