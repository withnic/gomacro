/*
 * gomacro - A Go intepreter with Lisp-like macros
 *
 * Copyright (C) 2017 Massimiliano Ghilardi
 *
 *     This program is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *
 *     This program is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *
 *     You should have received a copy of the GNU General Public License
 *     along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * output.go
 *
 *  Created on: Feb 19, 2017
 *      Author: Massimiliano Ghilardi
 */

package interpreter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	r "reflect"
	"sort"
)

type RuntimeError struct {
	FileSet
	Format string
	Args   []interface{}
}

func (err RuntimeError) Error() string {
	return fmt.Sprintf(err.Format, err.toPrintables(err.Args)...)
}

func Error(err error) interface{} {
	panic(err)
}

func Errorf(format string, args ...interface{}) {
	panic(RuntimeError{FileSet{nil}, format, args})
}

func (f FileSet) Errorf(format string, args ...interface{}) (r.Value, []r.Value) {
	panic(RuntimeError{f, format, args})
}

func (f FileSet) PackErrorf(format string, args ...interface{}) []r.Value {
	panic(RuntimeError{f, format, args})
}

func (p *Interpreter) Warnf(format string, args ...interface{}) {
	str := p.Sprintf(format, args...)
	fmt.Fprintf(p.Stderr, "warning: %s\n", str)
}

func (p *Interpreter) Debugf(format string, args ...interface{}) {
	str := p.Sprintf(format, args...)
	fmt.Fprintf(p.Stdout, "// debug: %s\n", str)
}

func BadIndex(index int, size int) AstWithNode {
	if size > 0 {
		Errorf("index out of range: %d not in 0...%d", index, size-1)
	} else {
		Errorf("index out of range: %d, slice is empty", index)
	}
	return nil
}

func (f FileSet) FprintValues(out io.Writer, values ...r.Value) {
	if len(values) == 0 {
		fmt.Fprint(out, "// no value\n")
		return
	}
	for _, v := range values {
		f.FprintValue(out, v)
	}
}

func (f FileSet) FprintValue(out io.Writer, v r.Value) {
	var vi interface{}
	var vt r.Type
	if v == None {
		fmt.Fprint(out, "// no value\n")
		return
	} else if v == Nil {
		vi = nil
		vt = nil
	} else {
		vi = v.Interface()
		vt = v.Type()
	}
	vi = f.toPrintable(vi)
	if vi == nil && vt == nil {
		fmt.Fprint(out, "<nil>\n")
		return
	}
	switch vi := vi.(type) {
	case uint, uint8, uint32, uint64, uintptr:
		fmt.Fprintf(out, "%d <%v>\n", vi, vt)
	default:
		if vt == typeOfString {
			fmt.Fprintf(out, "%#v <%v>\n", vi, vt)
		} else {
			fmt.Fprintf(out, "%v <%v>\n", vi, vt)
		}
	}
}

func (f FileSet) Fprintf(out io.Writer, format string, values ...interface{}) (n int, err error) {
	values = f.toPrintables(values)
	return fmt.Fprintf(out, format, values...)
}

func (f FileSet) Sprintf(format string, values ...interface{}) string {
	values = f.toPrintables(values)
	return fmt.Sprintf(format, values...)
}

func (f FileSet) toString(separator string, values ...interface{}) string {
	if len(values) == 0 {
		return ""
	}
	values = f.toPrintables(values)
	var buf bytes.Buffer
	for i, value := range values {
		if i != 0 {
			buf.WriteString(separator)
		}
		fmt.Fprint(&buf, value)
	}
	return buf.String()
}

func (f FileSet) toPrintables(values []interface{}) []interface{} {
	for i, vi := range values {
		values[i] = f.toPrintable(vi)
	}
	return values
}

func (f FileSet) toPrintable(value interface{}) (ret interface{}) {
	if value == nil {
		return nil
	}
	defer func() {
		if rec := recover(); rec != nil {
			ret = fmt.Sprintf("error pretty-printing %#v <%v>", value, r.TypeOf(value))
		}
	}()
	if v, ok := value.(r.Value); ok {
		return f.valueToPrintable(v)
	}
	v := r.ValueOf(value)
	k := v.Kind()
	if k == r.Array || k == r.Slice {
		n := v.Len()
		values := make([]interface{}, n)
		for i := 0; i < n; i++ {
			values[i] = f.toPrintable(v.Index(i))
		}
		return values
	}
	switch v := value.(type) {
	case AstWithNode:
		return f.nodeToPrintable(v.Node())
	case Ast:
		return f.toPrintable(v.Interface())
	case ast.Node:
		return f.nodeToPrintable(v)
	}
	return value
}

func (f FileSet) valueToPrintable(value r.Value) interface{} {
	if value == None {
		return "/*no value*/"
	} else if value == Nil {
		return nil
	} else {
		return f.toPrintable(value.Interface())
	}
}

func (f FileSet) nodeToPrintable(node ast.Node) interface{} {
	if node == nil {
		return nil
	}

	fset := f.Fileset
	if fset == nil {
		fset = token.NewFileSet()
	}
	var buf bytes.Buffer
	err := format.Node(&buf, fset, node)
	if err != nil {
		return err
	}
	return buf.String()
}

func (f FileSet) showHelp(out io.Writer) {
	fmt.Fprint(out, `// interpreter commands:
:env    show available functions, variables and constants
:help   show this help
:quit   quit the interpreter
`)
}

func (env *Env) showEnv(out io.Writer) {
	binds := env.Binds
	keys := make([]string, len(binds))
	i := 0
	for k := range binds {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	spaces15 := "               "
	for _, k := range keys {
		n := len(k) & 15
		fmt.Fprintf(out, "%s%s = ", k, spaces15[n:])
		env.FprintValue(out, binds[k])
	}
}