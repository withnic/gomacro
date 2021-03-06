/*
 * gomacro - A Go interpreter with Lisp-like macros
 *
 * Copyright (C) 2019 Massimiliano Ghilardi
 *
 *     This Source Code Form is subject to the terms of the Mozilla Public
 *     License, v. 2.0. If a copy of the MPL was not distributed with this
 *     file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 *
 * code.go
 *
 *  Created on Feb 10, 2019
 *      Author Massimiliano Ghilardi
 */

package jit

import (
	"github.com/cosmos72/gomacro/jit/asm"
)

// will be passed as argument to asm.Asm()
type Code []AsmCode

func (c *Code) Init() *Code {
	*c = nil
	return c
}

func (c *Code) Op1(op Op1, src Expr, dst SoftReg) *Code {
	*c = append(*c, op.Asm(), asmArg(src), dst.id)
	return c
}

func (c *Code) Op2(op Op2, a Expr, b Expr, dst SoftReg) *Code {
	*c = append(*c, op.Asm(), asmArg(a), asmArg(b), dst.id)
	return c
}

func (c *Code) Inst1(inst Inst1, dst Expr) *Code {
	if inst != NOP {
		*c = append(*c, inst.Asm(), asmArg(dst))
	}
	return c
}

func (c *Code) Inst2(inst Inst2, src Expr, dst Expr) *Code {
	asrc, adst := asmArg(src), asmArg(dst)
	if inst != ASSIGN || asrc != adst {
		*c = append(*c, inst.Asm(), asrc, adst)
	}
	return c
}

// ALLOC/FREE
func (c *Code) SoftReg(op asm.Op2Misc, id SoftRegId, kind Kind) SoftReg {
	*c = append(*c, op, id, kind)
	return SoftReg{id, kind}
}

func asmArg(e Expr) AsmCode {
	switch e := e.(type) {
	case AsmCode:
		return e
	case SoftReg:
		return e.id
	case *Expr1, *Expr2:
		errorf("internal error: cannot assemble %T, must be compiled first: %v", e, e)
	default:
		errorf("unknown Expr type %T: %v", e, e)
	}
	return nil
}
