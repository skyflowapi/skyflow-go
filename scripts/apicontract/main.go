// Command apicontract records the public API of a Go module as a text snapshot and
// compares two snapshots. It is the Go counterpart of the japicmp check in
// skyflow-java: the snapshots in api-report/ are the approved contract, and CI fails
// a pull request that removes or changes a line without approval.
//
// Usage:
//
//	apicontract snapshot <module-dir>          # print the snapshot to stdout
//	apicontract compare <old.api> <new.api>    # exit 1 when old lines are gone
//
// Each line is one fact about the API: a const and its value, a var, a func, a type,
// a struct field or a method. A line that disappears is a breaking change; a new line
// is an addition. Interfaces are one line, so adding a method to an interface (which
// breaks anyone implementing it) shows up as a change.
//
// The snapshot is alias-aware. Code moved to the common module stays in the v2 and
// flowvault packages as `type X = common/pkg.X`. Such an alias is written out as if
// the type were declared there, and every reference to the target type is written with
// the alias name, so a move that keeps the API the same produces the same lines.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/types"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const usage = `usage:
  apicontract snapshot <module-dir>
  apicontract compare <old.api> <new.api>`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		var breaking errBreaking
		if errors.As(err, &breaking) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "apicontract:", err)
		os.Exit(2)
	}
}

func run(args []string, out io.Writer) error {
	switch {
	case len(args) == 2 && args[0] == "snapshot":
		lines, err := snapshot(args[1])
		if err != nil {
			return err
		}
		for _, l := range lines {
			fmt.Fprintln(out, l)
		}
		return nil
	case len(args) == 3 && args[0] == "compare":
		return compare(args[1], args[2], out)
	default:
		return errors.New(usage)
	}
}

// errBreaking reports that a compare found removed or changed lines.
type errBreaking struct{ n int }

func (e errBreaking) Error() string { return fmt.Sprintf("%d breaking change(s)", e.n) }

// publicPackage reports whether a package is part of the module's public API.
func publicPackage(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if seg == "internal" || seg == "generated" {
			return false
		}
	}
	return true
}

func snapshot(dir string) ([]string, error) {
	cfg := &packages.Config{
		// Deps are type-checked from source so the tool does not depend on the export
		// data format of the Go toolchain that runs it.
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports |
			packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir: dir,
		// Each module is read with the dependency versions pinned in its own go.mod,
		// as consumers get it.
		Env: append(os.Environ(), "GOWORK=off"),
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}
	var public []*types.Package
	var loadErrs []string
	for _, p := range pkgs {
		for _, e := range p.Errors {
			loadErrs = append(loadErrs, e.Error())
		}
		if publicPackage(p.PkgPath) && p.Types != nil {
			public = append(public, p.Types)
		}
	}
	if len(loadErrs) > 0 {
		return nil, fmt.Errorf("loading %s:\n%s", dir, strings.Join(loadErrs, "\n"))
	}
	sort.Slice(public, func(i, j int) bool { return public[i].Path() < public[j].Path() })

	w := newWriter(public)
	for _, p := range public {
		w.pkg(p)
	}
	sort.Strings(w.lines)
	return w.lines, nil
}

type writer struct {
	// aliases maps "targetPkg.Target" to "aliasPkg.Alias" for aliases whose target
	// lives outside the module's public packages (common, internal packages).
	aliases map[string]string
	lines   []string
}

func newWriter(public []*types.Package) *writer {
	isPublic := map[string]bool{}
	for _, p := range public {
		isPublic[p.Path()] = true
	}
	w := &writer{aliases: map[string]string{}}
	for _, p := range public {
		scope := p.Scope()
		for _, name := range scope.Names() {
			tn, ok := scope.Lookup(name).(*types.TypeName)
			if !ok || !tn.Exported() || !tn.IsAlias() {
				continue
			}
			named, ok := types.Unalias(tn.Type()).(*types.Named)
			if !ok || named.Obj().Pkg() == nil || isPublic[named.Obj().Pkg().Path()] {
				continue
			}
			key := named.Obj().Pkg().Path() + "." + named.Obj().Name()
			val := p.Path() + "." + name
			if cur, ok := w.aliases[key]; !ok || val < cur {
				w.aliases[key] = val
			}
		}
	}
	return w
}

var qualifiedName = regexp.MustCompile(`[\w.~/-]+\.[A-Za-z_]\w*`)

// typ renders a type with full package paths, then swaps alias targets for the alias.
func (w *writer) typ(t types.Type) string {
	s := types.TypeString(t, func(p *types.Package) string { return p.Path() })
	if len(w.aliases) == 0 {
		return s
	}
	return qualifiedName.ReplaceAllStringFunc(s, func(m string) string {
		if a, ok := w.aliases[m]; ok {
			return a
		}
		return m
	})
}

// sig renders a signature without parameter names, which callers never depend on.
func (w *writer) sig(s *types.Signature) string {
	var b strings.Builder
	if tp := s.TypeParams(); tp != nil && tp.Len() > 0 {
		b.WriteString(w.typeParams(tp))
	}
	b.WriteString("(")
	params := s.Params()
	for i := 0; i < params.Len(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		t := params.At(i).Type()
		if s.Variadic() && i == params.Len()-1 {
			b.WriteString("..." + w.typ(t.(*types.Slice).Elem()))
			continue
		}
		b.WriteString(w.typ(t))
	}
	b.WriteString(")")
	res := s.Results()
	switch res.Len() {
	case 0:
	case 1:
		b.WriteString(" " + w.typ(res.At(0).Type()))
	default:
		b.WriteString(" (")
		for i := 0; i < res.Len(); i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(w.typ(res.At(i).Type()))
		}
		b.WriteString(")")
	}
	return b.String()
}

func (w *writer) typeParams(tp *types.TypeParamList) string {
	parts := make([]string, tp.Len())
	for i := 0; i < tp.Len(); i++ {
		p := tp.At(i)
		parts[i] = p.Obj().Name() + " " + w.typ(p.Constraint())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (w *writer) emit(pkg *types.Package, format string, args ...any) {
	w.lines = append(w.lines, "pkg "+pkg.Path()+", "+fmt.Sprintf(format, args...))
}

func (w *writer) pkg(p *types.Package) {
	scope := p.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}
		switch o := obj.(type) {
		case *types.Const:
			w.emit(p, "const %s %s = %s", name, w.typ(o.Type()), o.Val().ExactString())
		case *types.Var:
			w.emit(p, "var %s %s", name, w.typ(o.Type()))
		case *types.Func:
			w.emit(p, "func %s%s", name, w.sig(o.Type().(*types.Signature)))
		case *types.TypeName:
			w.typeName(p, o)
		}
	}
}

func (w *writer) typeName(p *types.Package, o *types.TypeName) {
	name := o.Name()
	t := o.Type()
	if o.IsAlias() {
		named, ok := types.Unalias(t).(*types.Named)
		if !ok {
			// An alias of an unnamed type (type X = string) is not a distinct type, so
			// it stays marked as an alias.
			w.emit(p, "type %s = %s", name, w.typ(types.Unalias(t)))
			return
		}
		t = named
	}
	named, ok := t.(*types.Named)
	if !ok {
		return
	}
	header := name
	if tp := named.TypeParams(); tp != nil && tp.Len() > 0 {
		header += w.typeParams(tp)
	}

	switch u := named.Underlying().(type) {
	case *types.Struct:
		w.emit(p, "type %s struct", header)
		for i := 0; i < u.NumFields(); i++ {
			f := u.Field(i)
			switch {
			case f.Embedded():
				w.emit(p, "type %s struct, embedded %s", name, w.typ(f.Type()))
			case f.Exported():
				w.emit(p, "type %s struct, field %s %s", name, f.Name(), w.typ(f.Type()))
			}
		}
	case *types.Interface:
		w.emit(p, "type %s interface { %s }", header, w.interfaceMethods(u))
		return // the methods are on the header line
	default:
		w.emit(p, "type %s %s", header, w.typ(u))
	}

	for _, recv := range []types.Type{named, types.NewPointer(named)} {
		ms := types.NewMethodSet(recv)
		recvName := name
		if _, ptr := recv.(*types.Pointer); ptr {
			recvName = "*" + name
		}
		for i := 0; i < ms.Len(); i++ {
			m := ms.At(i).Obj()
			if !m.Exported() {
				continue
			}
			// The pointer method set repeats the value methods; list each once.
			if recvName[0] == '*' && types.NewMethodSet(named).Lookup(m.Pkg(), m.Name()) != nil {
				continue
			}
			w.emit(p, "method (%s) %s%s", recvName, m.Name(), w.sig(m.Type().(*types.Signature)))
		}
	}
}

func (w *writer) interfaceMethods(it *types.Interface) string {
	var ms []string
	unexported := false
	for i := 0; i < it.NumMethods(); i++ {
		m := it.Method(i)
		if !m.Exported() {
			unexported = true
			continue
		}
		ms = append(ms, m.Name()+w.sig(m.Type().(*types.Signature)))
	}
	sort.Strings(ms)
	if unexported {
		// Callers outside the module cannot implement it, so adding methods is safe.
		ms = append(ms, "// has unexported methods")
	}
	for i := 0; i < it.NumEmbeddeds(); i++ {
		if _, isIface := it.EmbeddedType(i).Underlying().(*types.Interface); !isIface {
			ms = append(ms, w.typ(it.EmbeddedType(i))) // type-set constraint
		}
	}
	return strings.Join(ms, "; ")
}

func readLines(path string) (map[string]bool, error) {
	set := map[string]bool{}
	f, err := os.Open(path) // #nosec G304 -- paths come from the CI script
	if errors.Is(err, os.ErrNotExist) {
		return set, nil // a module that did not exist before: everything is new
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if l := strings.TrimSpace(sc.Text()); l != "" {
			set[l] = true
		}
	}
	return set, sc.Err()
}

func diff(a, b map[string]bool) []string {
	var out []string
	for l := range a {
		if !b[l] {
			out = append(out, l)
		}
	}
	sort.Strings(out)
	return out
}

func compare(oldPath, newPath string, out io.Writer) error {
	oldSet, err := readLines(oldPath)
	if err != nil {
		return err
	}
	newSet, err := readLines(newPath)
	if err != nil {
		return err
	}
	removed, added := diff(oldSet, newSet), diff(newSet, oldSet)

	switch {
	case len(oldSet) == 0:
		fmt.Fprintf(out, "New module: %d API lines become the baseline.\n", len(newSet))
		return nil
	case len(removed) == 0 && len(added) == 0:
		fmt.Fprintln(out, "No API changes.")
		return nil
	}
	if len(removed) > 0 {
		fmt.Fprintf(out, "Breaking: %d line(s) removed or changed\n", len(removed))
		for _, l := range removed {
			fmt.Fprintln(out, "- "+l)
		}
	}
	if len(added) > 0 {
		fmt.Fprintf(out, "Added: %d line(s)\n", len(added))
		for _, l := range added {
			fmt.Fprintln(out, "+ "+l)
		}
	}
	if len(removed) > 0 {
		return errBreaking{len(removed)}
	}
	return nil
}
