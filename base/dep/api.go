/*
 * gomacro - A Go interpreter with Lisp-like macros
 *
 * Copyright (C) 2017-2018 Massimiliano Ghilardi
 *
 *     This program is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU Lesser General Public License as published
 *     by the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *
 *     This program is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU Lesser General Public License for more details.
 *
 *     You should have received a copy of the GNU Lesser General Public License
 *     along with this program.  If not, see <https://www.gnu.org/licenses/lgpl>.
 *
 *
 * api.go
 *
 *  Created on: May 03, 2018
 *      Author: Massimiliano Ghilardi
 */

package dep

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Support for out-of-order code

type Kind int

const (
	Const Kind = iota
	Func
	Import
	Macro
	Method
	Type
	TypeFwd
	Var
)

var kinds = map[Kind]string{
	Const:   "Const",
	Func:    "Func",
	Import:  "Import",
	Macro:   "Macro",
	Method:  "Method",
	Type:    "Type",
	TypeFwd: "TypeFwd", // forward type declaration
	Var:     "Var",
}

func (k Kind) String() string {
	name, ok := kinds[k]
	if ok {
		return name
	}
	return fmt.Sprintf("Kind%d", int(k))
}

// for multiple const or var declarations in a single *ast.ValueSpec
type Extra struct {
	Ident *ast.Ident
	Type  ast.Expr
	Value ast.Expr
}

type Decl struct {
	Kind  Kind
	Name  string
	Node  ast.Node // nil for multiple const or var declarations in a single *ast.ValueSpec - in such case, see Extra
	Deps  []string // names of types, constants and variables used in Node's declaration
	Pos   token.Pos
	Iota  int // for constants, value of iota to use
	Extra *Extra
}

type DeclMap map[string]*Decl

type DeclList []*Decl

type Scope struct {
	Decls  DeclMap
	Outer  *Scope
	Gensym int
}

func NewLoader() *Scope {
	return &Scope{
		Decls: make(map[string]*Decl),
	}
}

type Sorter struct {
	Loader Scope
}

func NewSorter() *Sorter {
	return &Sorter{
		Loader: Scope{
			Decls: make(map[string]*Decl),
		},
	}
}

// Sorter resolves top-level constant, type, function and var
// declaration order by analyzing their dependencies.
//
// also resolves top-level var initialization order
// analyzing their dependencies.