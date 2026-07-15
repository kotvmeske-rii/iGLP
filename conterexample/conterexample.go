package conterexample

import (
	"context"
	"iglp/solver"
	"iglp/syntax"
)

// глобальный, хранит итоговый граф
type Contermodel struct {
	count int
	Frame *solver.KripkeFrame
}

type Trace struct {
	WorldNumber int
	Child       []*Trace
	Val         map[solver.FormulaNumber]bool
}

// создаем наш изначальный нулевой мир
func NewContermodel() *Contermodel {
	return &Contermodel{
		count: 0,
		Frame: &solver.KripkeFrame{
			Worlds:    make(map[int]*solver.ModelWorld),
			Relations: make(map[int][]int),
		},
	}
}

// генерация айди мира
func (c *Contermodel) NextWorldNumber() int {
	number := c.count
	c.count++

	return number
}

// Проверяет текущий мир на наличие явного логического противоречия
// true - одна и та же формула одновременно находится в списках истинных и ложных
func (c *Contermodel) Denial(world *solver.ModelWorld) bool {
	for _, v := range world.TrueFormula {
		for _, m := range world.FalseFormula {
			if v == m {
				return true
			}
		}
	}

	return false
}

// true - контпример есть, те есть хотя бы одна открытая ветка,
// false - контрпримера нет, те все ветки закрыты(логическое противоречие)
func (c *Contermodel) Prove(
	ctx context.Context,
	worldNumber int,
	t int,
	f int,
) (*Trace, bool) {
	world := c.Frame.Worlds[worldNumber]

	// Step 1: проверка на логическое противоречие
	if c.Denial(world) {
		return nil, false //невозможно чтобы формула в мире была одновременно и верна и не верна
	}

	bibliothek := solver.GetContext(ctx)

	// Step 2: local world

	// constants
	// проверка противоречий с константой ложь и константой истина
	if c.constants(bibliothek, &world.TrueFormula, syntax.TokFalse) ||
		c.constants(bibliothek, &world.FalseFormula, syntax.TokTrue) {
		return nil, false
	}

	if t < len(world.TrueFormula) {
		v := world.TrueFormula[t]
		key := bibliothek.Key(v)

		switch key.Op {
		// Удаление тривиальной константы истина
		case syntax.TokTrue:
			trace, boolean := c.proveConst(
				ctx, worldNumber, t, f,
				&world.TrueFormula, 1, 0,
			)

			return trace, boolean
		// true impl
		case syntax.TokImpl:
			trace, boolean := c.proveTrueImpl(
				key, ctx, worldNumber, t, f,
				&world.FalseFormula, &world.TrueFormula,
			)

			return trace, boolean
		// variate true
		case syntax.TokVar:
			trace, boolean := c.proveVal(
				ctx, worldNumber, t, f,
				world.Valuation, 1, 0, v, true,
			)

			return trace, boolean
		default:
		}

		return c.Prove(ctx, worldNumber, t+1, f)
	}

	if f < len(world.FalseFormula) {
		v := world.FalseFormula[f]
		key := bibliothek.Key(v)

		switch key.Op {
		// Удаление тривиальной константы ложь
		case syntax.TokFalse:
			trace, boolean := c.proveConst(
				ctx, worldNumber, t, f,
				&world.FalseFormula, 0, 1,
			)

			return trace, boolean
		// false impl
		// Импликация ложна iff A истинно,22 B ложно
		case syntax.TokImpl:
			trace, boolean := c.proveFalseImpl(
				key, ctx, worldNumber, t, f,
				&world.FalseFormula, &world.TrueFormula,
			)

			return trace, boolean
		// var false
		case syntax.TokVar:
			trace, boolean := c.proveVal(
				ctx, worldNumber, t, f,
				world.Valuation, 0, 1, v, false,
			)

			return trace, boolean
		default:
			return c.Prove(ctx, worldNumber, t, f+1)
		}
	}

	var children []*Trace

	// Step 3: next world
	// false box
	trace, boolean := c.proveBox(
		ctx, children, &world.FalseFormula, &world.TrueFormula,
		bibliothek, world.Valuation, world.Number,
	)

	return trace, boolean
}

func (c *Contermodel) Sammeln(root *Trace) {
	if root == nil {
		return
	}

	c.Frame.Worlds[root.WorldNumber] = &solver.ModelWorld{
		Number:    root.WorldNumber,
		Valuation: root.Val,
	}

	for _, child := range root.Child {
		c.Frame.Relations[root.WorldNumber] = append(
			c.Frame.Relations[root.WorldNumber],
			child.WorldNumber,
		)
		c.Sammeln(child)
	}
}

// Экспортирует построенную шкалу Крипке
func (c *Contermodel) InputToKripke() *solver.KripkeFrame {
	frame := &solver.KripkeFrame{
		Worlds:    make(map[int]*solver.ModelWorld),
		Relations: c.Frame.Relations,
	}

	for k, v := range c.Frame.Worlds {
		frame.Worlds[k] = &solver.ModelWorld{
			Number:       v.Number,
			Valuation:    v.Valuation,
			TrueFormula:  []solver.FormulaNumber{},
			FalseFormula: []solver.FormulaNumber{},
		}
	}

	return frame
}

func (c *Contermodel) constants(
	bibliothek *solver.FormulaBibliothek,
	formula *[]solver.FormulaNumber,
	constant syntax.TokenType,
) bool {
	for _, v := range *formula {
		if bibliothek.Key(v).Op == constant {
			return true
		}
	}

	return false
}

func (c *Contermodel) proveConst(
	ctx context.Context,
	worldNumber int,
	t int,
	f int,
	formula *[]solver.FormulaNumber,
	indT int,
	indF int,
) (*Trace, bool) {
	lang := len(*formula)
	trace, boolean := c.Prove(ctx, worldNumber, t+indT, f+indF)
	*formula = (*formula)[:lang]

	return trace, boolean
}

// true impl
// Ветвление на два альтернативных сценария: не A или B
func (c *Contermodel) proveTrueImpl(
	key solver.FormulaKey,
	ctx context.Context,
	worldNumber int,
	t int, f int,
	formula_false *[]solver.FormulaNumber,
	formula_true *[]solver.FormulaNumber,
) (*Trace, bool) {
	leftKey := key.Left
	rightKey := key.Right

	// left false
	langLeft := len(*formula_false)

	*formula_false = append(*formula_false, leftKey)

	if trace, boolean := c.Prove(
		ctx, worldNumber, t+1, f,
	); boolean {
		return trace, true
	}

	*formula_false = (*formula_false)[:langLeft]

	// right true
	langRight := len(*formula_true)
	*formula_true = append(*formula_true, rightKey)
	trace, boolean := c.Prove(ctx, worldNumber, t+1, f)
	*formula_true = (*formula_true)[:langRight]

	return trace, boolean
}

// false impl
// Импликация ложна iff A истинно,22 B ложно
func (c *Contermodel) proveFalseImpl(
	key solver.FormulaKey,
	ctx context.Context,
	worldNumber int,
	t int, f int,
	formula_false *[]solver.FormulaNumber,
	formula_true *[]solver.FormulaNumber,
) (*Trace, bool) {
	leftKey := key.Left
	rightKey := key.Right
	langLeft := len(*formula_true)
	langRight := len(*formula_false)

	*formula_true = append(*formula_true, leftKey)
	*formula_false = append(*formula_false, rightKey)

	trace, boolean := c.Prove(ctx, worldNumber, t, f+1)

	*formula_true = (*formula_true)[:langLeft]
	*formula_false = (*formula_false)[:langRight]

	return trace, boolean
}

func (c *Contermodel) proveVal(
	ctx context.Context,
	worldNumber int,
	t int,
	f int,
	formulaVal map[solver.FormulaNumber]bool,
	indT int,
	indF int,
	v solver.FormulaNumber,
	truth bool,
) (*Trace, bool) {
	key, value := formulaVal[v]
	formulaVal[v] = truth
	trace, boolean := c.Prove(
		ctx, worldNumber, t+indT, f+indF,
	)

	if value {
		formulaVal[v] = key
	} else {
		delete(formulaVal, v)
	}

	return trace, boolean
}

func (c *Contermodel) proveBox(
	ctx context.Context,
	children []*Trace,
	formulaT *[]solver.FormulaNumber,
	formulaF *[]solver.FormulaNumber,
	bibliothek *solver.FormulaBibliothek,
	formulaVal map[solver.FormulaNumber]bool,
	formulaNumber int,
) (*Trace, bool) {
	// false box
	for _, v := range *formulaF {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokBox {
			childNumber := key.Left

			// Новый дотижимый мир
			newNumber := c.NextWorldNumber()
			newWorld := NewModelWorld(newNumber)
			newWorld.FalseFormula = append(newWorld.FalseFormula, childNumber, v)

			// if in w_0 []B - true => in w_1 B, []B - true
			for _, va := range *formulaT {
				trueKey := bibliothek.Key(va)

				if trueKey.Op == syntax.TokBox {
					newWorld.TrueFormula = append(newWorld.TrueFormula, trueKey.Left, va)
				}
			}

			newWorld.TrueFormula = append(newWorld.TrueFormula, v)

			kind, ok := c.Prove(ctx, newNumber, 0, 0)

			if !ok {
				return nil, false
			}
			children = append(children, kind)
		}
	}

	continueVal := make(map[solver.FormulaNumber]bool)

	for k, value := range formulaVal {
		continueVal[k] = value
	}

	return &Trace{
		WorldNumber: formulaNumber,
		Child:       children,
		Val:         continueVal,
	}, true
}

//Разносим импликации в 2 разные формулы, у них слишком разная логика
