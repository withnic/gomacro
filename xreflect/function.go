/*
 * gomacro - A Go interpreter with Lisp-like macros
 *
 * Copyright (C) 2017 Massimiliano Ghilardi
 *
 *     This program is free software you can redistribute it and/or modify
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
 *     along with this program.  If not, see <http//www.gnu.org/licenses/>.
 *
 * type.go
 *
 *  Created on May 07, 2017
 *      Author Massimiliano Ghilardi
 */

package xreflect

import (
	"go/types"
	"reflect"
)

// IsMethod reports whether a function type's contains a receiver, i.e. is a method.
// It panics if the type's Kind is not Func.
func (t *xtype) IsMethod() bool {
	if t.Kind() != reflect.Func {
		xerrorf(t, "IsMethod of non-func type %v", t)
	}
	gtype := t.gtype.(*types.Signature)
	return gtype.Recv() != nil
}

// IsVariadic reports whether a function type's final input parameter is a "..." parameter.
// If so, t.In(t.NumIn() - 1) returns the parameter's implicit actual type []T.
// IsVariadic panics if the type's Kind is not Func.
func (t *xtype) IsVariadic() bool {
	if t.Kind() != reflect.Func {
		xerrorf(t, "In of non-func type %v", t)
	}
	return t.rtype.IsVariadic()
}

// In returns the type of a function type's i'th input parameter.
// It panics if the type's Kind is not Func.
// It panics if i is not in the range [0, NumIn()).
func (t *xtype) In(i int) Type {
	if t.Kind() != reflect.Func {
		xerrorf(t, "In of non-func type %v", t)
	}
	gtype := t.underlying().(*types.Signature)
	va := gtype.Params().At(i)
	if gtype.Recv() != nil {
		i++ // skip the receiver in reflect.Type
	}
	return t.universe.MakeType(va.Type(), t.rtype.In(i))
}

// NumIn returns a function type's input parameter count.
// It panics if the type's Kind is not Func.
func (t *xtype) NumIn() int {
	if t.Kind() != reflect.Func {
		xerrorf(t, "NumIn of non-func type %v", t)
	}
	gtype := t.underlying().(*types.Signature)
	return gtype.Params().Len()
}

// NumOut returns a function type's output parameter count.
// It panics if the type's Kind is not Func.
func (t *xtype) NumOut() int {
	if t.Kind() != reflect.Func {
		xerrorf(t, "NumOut of non-func type %v", t)
	}
	gtype := t.underlying().(*types.Signature)
	return gtype.Results().Len()
}

// Out returns the type of a function type's i'th output parameter.
// It panics if the type's Kind is not Func.
// It panics if i is not in the range [0, NumOut()).
func (t *xtype) Out(i int) Type {
	if t.Kind() != reflect.Func {
		xerrorf(t, "Out of non-func type %v", t)
	}
	gtype := t.underlying().(*types.Signature)
	va := gtype.Results().At(i)
	return t.universe.MakeType(va.Type(), t.rtype.Out(i))
}

// Recv returns the type of a method type's receiver parameter.
// It panics if the type's Kind is not Func.
// It returns nil if t has no receiver.
func (t *xtype) Recv() Type {
	if t.Kind() != reflect.Func {
		xerrorf(t, "Recv of non-func type %v", t)
	}
	gtype := t.underlying().(*types.Signature)
	va := gtype.Recv()
	if va == nil {
		return nil
	}
	return t.universe.MakeType(va.Type(), t.rtype.In(0))
}

func FuncOf(in []Type, out []Type, variadic bool) Type {
	return MethodOf(nil, in, out, variadic)
}

func (v *Universe) FuncOf(in []Type, out []Type, variadic bool) Type {
	return v.MethodOf(nil, in, out, variadic)
}

func MethodOf(recv Type, in []Type, out []Type, variadic bool) Type {
	v := universe
	if recv != nil {
		v = recv.Universe()
	} else if len(in) != 0 && in[0] != nil {
		v = in[0].Universe()
	} else if len(out) != 0 && out[0] != nil {
		v = out[0].Universe()
	}
	return v.MethodOf(recv, in, out, variadic)
}

func (v *Universe) MethodOf(recv Type, in []Type, out []Type, variadic bool) Type {
	gin := toGoTuple(in)
	gout := toGoTuple(out)
	rin := toReflectTypes(in)
	rout := toReflectTypes(out)
	var grecv *types.Var
	if recv != nil {
		rin = append([]reflect.Type{recv.ReflectType()}, rin...)
		grecv = toGoParam(recv)
	}
	gtype := types.NewSignature(grecv, gin, gout, variadic)
	if grecv != nil {
		debugf("xreflect.MethodOf: recv = <%v>, method = <%v> with recv = <%v>", grecv, gtype, gtype.Recv())
	}
	return v.MakeType(
		gtype,
		reflect.FuncOf(rin, rout, variadic),
	)
}
