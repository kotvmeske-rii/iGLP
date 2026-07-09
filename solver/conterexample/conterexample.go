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
func (c *Contermodel) Prove(ctx context.Context, worldNumber int, t int, f int) (*Trace, bool) {
	world := c.Frame.Worlds[worldNumber]

	//Step 1: проверка на логическое противоречие
	if c.Denial(world) {
		return nil, false //невозможно чтобы формула в мире была одновременно и верна и не верна
	}

	bibliothek := solver.GetContext(ctx)

	//Step 2: local world

	//constants
	//проверка противоречий с константой ложь
	for _, v := range world.TrueFormula {
		if bibliothek.Key(v).Op == syntax.TokFalse {
			return nil, false
		}
	}

	//проверка противоречий с константой истина
	for _, v := range world.FalseFormula {
		if bibliothek.Key(v).Op == syntax.TokTrue {
			return nil, false
		}
	}

	if t < len(world.TrueFormula) {
		v := world.TrueFormula[t]
		key := bibliothek.Key(v)

		switch key.Op {
		// Удаление тривиальной константы истина
		case syntax.TokTrue:
			lang := len(world.TrueFormula)
			trace, boolean := c.Prove(ctx, worldNumber, t+1, f)
			world.TrueFormula = world.TrueFormula[:lang]
			return trace, boolean
		//true impl
		//Ветвление на два альтернативных сценария: не A или B
		case syntax.TokImpl:
			leftKey := key.Left
			rightKey := key.Right

			//left false
			langLeft := len(world.FalseFormula)
			world.FalseFormula = append(world.FalseFormula, leftKey)
			if trace, boolean := c.Prove(ctx, worldNumber, t+1, f); boolean {
				return trace, true
			}

			world.FalseFormula = world.FalseFormula[:langLeft]

			//right true
			langRight := len(world.TrueFormula)
			world.TrueFormula = append(world.TrueFormula, rightKey)
			trace, boolean := c.Prove(ctx, worldNumber, t+1, f)
			world.TrueFormula = world.TrueFormula[:langRight]

			return trace, boolean
		//variate true
		case syntax.TokVar:
			key, value := world.Valuation[v]
			world.Valuation[v] = true
			trace, boolean := c.Prove(ctx, worldNumber, t+1, f)

			if value {
				world.Valuation[v] = key
			} else {
				delete(world.Valuation, v)
			}

			return trace, boolean
		}

		return c.Prove(ctx, worldNumber, t+1, f)
	}

	if f < len(world.FalseFormula) {
		v := world.FalseFormula[f]
		key := bibliothek.Key(v)

		switch key.Op {
		// Удаление тривиальной константы ложь
		case syntax.TokFalse:
			lang := len(world.FalseFormula)
			trace, boolean := c.Prove(ctx, worldNumber, t, f+1)
			world.FalseFormula = world.FalseFormula[:lang]
			return trace, boolean
		//false impl
		// Импликация ложна iff A истинно,22 B ложно
		case syntax.TokImpl:
			leftKey := key.Left
			rightKey := key.Right
			langLeft := len(world.TrueFormula)
			langRight := len(world.FalseFormula)

			world.TrueFormula = append(world.TrueFormula, leftKey)
			world.FalseFormula = append(world.FalseFormula, rightKey)

			trace, boolean := c.Prove(ctx, worldNumber, t, f+1)

			world.TrueFormula = world.TrueFormula[:langLeft]
			world.FalseFormula = world.FalseFormula[:langRight]

			return trace, boolean
		//var false
		case syntax.TokVar:
			key, value := world.Valuation[v]
			world.Valuation[v] = false
			trace, boolean := c.Prove(ctx, worldNumber, t, f+1)

			if value {
				world.Valuation[v] = key
			} else {
				delete(world.Valuation, v)
			}

			return trace, boolean
		default:
			return c.Prove(ctx, worldNumber, t, f+1)
		}
	}

	var children []*Trace

	//Step 3: next world
	//false box
	for _, v := range world.FalseFormula {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokBox {
			childNumber := key.Left

			// Новый дотижимый мир
			newNumber := c.NextWorldNumber()
			newWorld := NewModelWorld(newNumber)
			newWorld.FalseFormula = append(newWorld.FalseFormula, childNumber, v)

			//if in w_0 []B - true => in w_1 B, []B - true
			for _, va := range world.TrueFormula {
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

	for k, value := range world.Valuation {
		continueVal[k] = value
	}

	return &Trace{
		WorldNumber: world.Number,
		Child:       children,
		Val:         continueVal,
	}, true
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
		c.Frame.Relations[root.WorldNumber] = append(c.Frame.Relations[root.WorldNumber], child.WorldNumber)
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
			Number:    v.Number,
			Valuation: v.Valuation,
		}
	}

	return frame
}
