// Command internalcheck fails when a module's public API exposes a type from one of the
// module's internal packages.
//
// apidiff compares public packages only. Users can't import internal packages, so that is
// safe while no internal type is reachable from the public API: a public alias of an
// internal type, an exported field, a function parameter or result, or a method of an
// exported type. This check keeps it that way.
//
// Usage:
//
//	internalcheck <module-dir>
//
// Exit status: 0 = no exposure, 1 = exposures found (listed on stdout), 2 = error.
package main

import (
	"errors"
	"fmt"
	"go/types"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// errExposed reports that the public API exposes internal types.
var errExposed = errors.New("public API exposes internal types")

func main() {
	os.Exit(exitCode(run(os.Args[1:], os.Stdout), os.Stderr))
}

// exitCode maps the result of run to the exit status, printing any tool error.
func exitCode(err error, stderr io.Writer) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errExposed):
		return 1
	default:
		fmt.Fprintln(stderr, "internalcheck:", err)
		return 2
	}
}

func run(args []string, out io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: internalcheck <module-dir>")
	}
	found, err := exposures(args[0])
	if err != nil {
		return err
	}
	for _, f := range found {
		fmt.Fprintln(out, f)
	}
	if len(found) > 0 {
		return errExposed
	}
	return nil
}

// isInternal reports whether an import path is an internal package or under one.
func isInternal(path string) bool {
	return strings.HasSuffix(path, "/internal") || strings.Contains(path, "/internal/")
}

// exposures lists every exported name in the module's public packages whose type reaches
// a type declared in an internal package, as "<name> -> <internal type>".
func exposures(dir string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps,
		Dir:  dir,
		Env:  append(os.Environ(), "GOWORK=off"),
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}
	var loadErrs []string
	found := map[string]bool{}
	for _, p := range pkgs {
		for _, e := range p.Errors {
			loadErrs = append(loadErrs, e.Error())
		}
		if p.Types == nil || p.Name == "main" || isInternal(p.PkgPath) {
			continue
		}
		scope := p.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if !obj.Exported() {
				continue
			}
			w := walker{pkg: p.Types, root: obj, seen: map[types.Type]bool{}, found: found}
			w.walk(p.PkgPath+"."+name, obj.Type())
		}
	}
	if len(loadErrs) > 0 {
		return nil, fmt.Errorf("loading %s:\n%s", dir, strings.Join(loadErrs, "\n"))
	}
	out := make([]string, 0, len(found))
	for f := range found {
		out = append(out, f)
	}
	sort.Strings(out)
	return out, nil
}

// walker follows one exported name through every type a user can reach from it.
type walker struct {
	pkg   *types.Package
	root  types.Object
	seen  map[types.Type]bool
	found map[string]bool
}

func (w *walker) walk(where string, t types.Type) {
	if t == nil || w.seen[t] {
		return
	}
	w.seen[t] = true
	switch u := t.(type) {
	case *types.Alias:
		w.walk(where, types.Unalias(u))
	case *types.Named:
		w.named(where, u)
	case *types.Pointer:
		w.walk(where, u.Elem())
	case *types.Slice:
		w.walk(where, u.Elem())
	case *types.Array:
		w.walk(where, u.Elem())
	case *types.Chan:
		w.walk(where, u.Elem())
	case *types.Map:
		w.walk(where, u.Key())
		w.walk(where, u.Elem())
	case *types.Signature:
		w.tuple(where, u.Params())
		w.tuple(where, u.Results())
	case *types.Struct:
		for i := 0; i < u.NumFields(); i++ {
			// Fields of an embedded struct are promoted, so an embedded field counts
			// even when its own name is unexported.
			if f := u.Field(i); f.Exported() || f.Embedded() {
				w.walk(where+"."+f.Name(), f.Type())
			}
		}
	case *types.Interface:
		for i := 0; i < u.NumMethods(); i++ {
			m := u.Method(i)
			w.walk(where+"."+m.Name(), m.Type())
		}
		for i := 0; i < u.NumEmbeddeds(); i++ {
			w.walk(where, u.EmbeddedType(i))
		}
	case *types.Union:
		for i := 0; i < u.Len(); i++ {
			w.walk(where, u.Term(i).Type())
		}
	case *types.TypeParam:
		w.walk(where, u.Constraint())
	}
}

func (w *walker) named(where string, n *types.Named) {
	if args := n.TypeArgs(); args != nil {
		for i := 0; i < args.Len(); i++ {
			w.walk(where, args.At(i))
		}
	}
	obj := n.Obj()
	if obj.Pkg() == nil {
		return // predeclared, such as error
	}
	if isInternal(obj.Pkg().Path()) {
		// The first internal type on a path is the leak to fix; what lies behind it
		// is only reachable through it.
		w.found[fmt.Sprintf("%s -> %s.%s", where, obj.Pkg().Path(), obj.Name())] = true
		return
	}
	// A type from another public package is checked when that package is, one from
	// another module is that module's own contract, and another exported type of this
	// package is checked as its own name.
	if obj.Pkg() != w.pkg || (obj.Exported() && obj != w.root) {
		return
	}
	w.walk(where, n.Underlying())
	for i := 0; i < n.NumMethods(); i++ {
		if m := n.Method(i); m.Exported() {
			w.walk(where+"."+m.Name(), m.Type())
		}
	}
}

func (w *walker) tuple(where string, t *types.Tuple) {
	for i := 0; i < t.Len(); i++ {
		w.walk(where, t.At(i).Type())
	}
}
