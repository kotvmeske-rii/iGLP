package syntax

import "fmt"

type Node interface {
	Type() TokenType
	String() string // Возвращает строковое представление формулы
}

// Логические константы
type VarConst struct {
	Op string
}

func (v *VarConst) Type() TokenType {
	if v.Op == "_" {
		return TokFalse
	}
	return TokTrue
}

func (v *VarConst) String() string {
	return v.Op
}

// Пропозициональная переменная
type VarNode struct {
	Name string
}

func (v *VarNode) Type() TokenType {
	return TokVar
}

func (v *VarNode) String() string {
	return v.Name
}

// Унарные операторы(отрицание и box)
type UnaryNode struct {
	Op    string // [] or ~
	Child Node   // подформула к которой применяем оператор
}

func (u *UnaryNode) Type() TokenType {
	if u.Op == "~" {
		return TokNot
	}
	return TokBox
}

func (u *UnaryNode) GetChild() Node {
	return u.Child
}

func (u *UnaryNode) String() string {
	return fmt.Sprintf("%s(%s)", u.Op, u.Child.String())
}

//Бинарные операции(и, или, импликация)
type BinaryNode struct {
	Op    string
	Left  Node
	Right Node
}

func (b *BinaryNode) Type() TokenType {
	switch b.Op {
	case "&":
		return TokAnd
	case "|":
		return TokOr
	default:
		return TokImpl
	}
}

func (b *BinaryNode) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Op, b.Right.String())
}

func (b *BinaryNode) GetLeft() Node {
	return b.Left
}

func (b *BinaryNode) GetRight() Node {
	return b.Right
}
