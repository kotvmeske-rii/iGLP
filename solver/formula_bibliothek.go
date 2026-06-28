package solver

import (
	"context"
	"iglp/syntax"
)

// Числовой идентификатор формулы
type FormulaNumber int

type ctxKey struct{}

type FormulaKey struct {
	Op    syntax.TokenType // Тип токена
	Name  string           // только для TokVar
	Left  FormulaNumber    // Единственный для унарных операций
	Right FormulaNumber
}

// Преобразуем структуру ast в числовые ID
// Можем сравнивать формулы за О(1)
type FormulaBibliothek struct {
	keyToNumber  map[FormulaKey]FormulaNumber
	numberToKey  map[FormulaNumber]FormulaKey
	numberToNode map[FormulaNumber]syntax.Node
	nextNumber   FormulaNumber
}

func NewFormulaBibliothek() *FormulaBibliothek {
	return &FormulaBibliothek{
		keyToNumber:  make(map[FormulaKey]FormulaNumber),
		numberToKey:  make(map[FormulaNumber]FormulaKey),
		numberToNode: make(map[FormulaNumber]syntax.Node),
		nextNumber:   1, // 0 <-> nil
	}
}

// рекурсивно перекидываем узел в кэш библиотеки
func (f *FormulaBibliothek) Bibliothek(node syntax.Node) FormulaNumber {
	if node == nil {
		return 0
	}

	var key FormulaKey
	key.Op = node.Type()

	switch n := node.(type) {
	case *syntax.VarNode:
		key.Name = n.Name
	case *syntax.UnaryNode:
		key.Left = f.Bibliothek(n.GetChild()) //рекурсивно разбираем подформулу
	case *syntax.BinaryNode:
		key.Left = f.Bibliothek(n.GetLeft())
		key.Right = f.Bibliothek(n.GetRight())
	case *syntax.VarConst:
		key.Name = n.Op
	}

	// Если формула встречалась раньше, возвращаем её ID
	if number, ext := f.keyToNumber[key]; ext {
		return number
	}

	//if new - generate number
	number := f.nextNumber
	f.nextNumber++

	f.keyToNumber[key] = number
	f.numberToKey[number] = key
	f.numberToNode[number] = node

	return number
}

// Возвращает FormulaKey за O(1)
func (f *FormulaBibliothek) Key(number FormulaNumber) FormulaKey {
	return f.numberToKey[number]
}

// Возвращает исходный узел за О(1)
func (f *FormulaBibliothek) Node(number FormulaNumber) syntax.Node {
	return f.numberToNode[number]
}

func PackBibliothek(ctx context.Context, bibl *FormulaBibliothek) context.Context {
	return context.WithValue(ctx, ctxKey{}, bibl)
}

func GetContext(ctx context.Context) *FormulaBibliothek {
	if bibl, ok := ctx.Value(ctxKey{}).(*FormulaBibliothek); ok {
		return bibl
	}

	return nil
}
