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
func (c *Contermodel) Prove(ctx context.Context, worldNumber int, t int, f int, trace map[solver.FormulaNumber]struct{}) bool {
	world := c.Frame.Worlds[worldNumber]

	//Step 1: проверка на логическое противоречие
	if c.Denial(world) {
		return false //невозможно чтобы формула в мире была одновременно и верна и не верна
	}

	bibliothek := solver.GetContext(ctx)

	//Step 2: local world

	//constants
	//проверка противоречий с константой ложь
	for _, v := range world.TrueFormula {
		if bibliothek.Key(v).Op == syntax.TokFalse {
			return false
		}
	}

	//проверка противоречий с константой истина
	for _, v := range world.FalseFormula {
		if bibliothek.Key(v).Op == syntax.TokTrue {
			return false
		}
	}

	if t < len(world.TrueFormula) {
		v := world.TrueFormula[t]
		key := bibliothek.Key(v)

		switch key.Op {
		// Удаление тривиальной константы истина
		case syntax.TokTrue:
			return c.Prove(ctx, worldNumber, t+1, f, trace)
		//true impl
		//Ветвление на два альтернативных сценария: не A или B
		case syntax.TokImpl:
			leftKey := key.Left
			rightKey := key.Right

			//left false
			world.FalseFormula = append(world.FalseFormula, leftKey)
			if c.Prove(ctx, worldNumber, t+1, f, trace) {
				return true
			}

			world.FalseFormula = world.FalseFormula[:len(world.FalseFormula)-1]

			//right true
			world.TrueFormula = append(world.TrueFormula, rightKey)
			if c.Prove(ctx, worldNumber, t+1, f, trace) {
				return true
			}

			world.TrueFormula = world.TrueFormula[:len(world.TrueFormula)-1]

			return false
		//variate true
		case syntax.TokVar:
			world.Valuation[v] = true
			result := c.Prove(ctx, worldNumber, t+1, f, trace)
			return result
		default:
			return c.Prove(ctx, worldNumber, t+1, f, trace)
		}
	}

	if f < len(world.FalseFormula) {
		v := world.FalseFormula[f]
		key := bibliothek.Key(v)

		switch key.Op {
		// Удаление тривиальной константы ложь
		case syntax.TokFalse:
			return c.Prove(ctx, worldNumber, t, f+1, trace)
		//false impl
		// Импликация ложна iff A истинно,22 B ложно
		case syntax.TokImpl:
			leftKey := key.Left
			rightKey := key.Right

			world.TrueFormula = append(world.TrueFormula, leftKey)
			world.FalseFormula = append(world.FalseFormula, rightKey)

			result := c.Prove(ctx, worldNumber, t, f+1, trace)

			world.TrueFormula = world.TrueFormula[:len(world.TrueFormula)-1]
			world.FalseFormula = world.FalseFormula[:len(world.FalseFormula)-1]

			return result
		//var false
		case syntax.TokVar:
			world.Valuation[v] = false
			result := c.Prove(ctx, worldNumber, t, f+1, trace)
			return result
		default:
			return c.Prove(ctx, worldNumber, t, f+1, trace)
		}
	}

	box := false
	endAllBox := true

	continueWorld := make([]int, 0)
	//Step 3: next world
	//false box
	for _, v := range world.FalseFormula {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokBox {
			box = true
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

			c.Frame.Worlds[newNumber] = newWorld
			c.Frame.Relations[worldNumber] = append(c.Frame.Relations[worldNumber], newNumber)
			for parent, child := range c.Frame.Relations {
				for _, ch := range child {
					if ch == worldNumber {
						c.Frame.Relations[parent] = append(c.Frame.Relations[parent], newNumber)
					}
				}
			}

			if !c.Prove(ctx, newNumber, 0, 0, trace) {
				endAllBox = false
				break
			}
		}
	}

	if box {
		if endAllBox {
			return true
		}

		for _, v := range continueWorld {
			delete(c.Frame.Worlds, v)
		}

		c.Frame.Relations[worldNumber] = nil

		return false
	}

	return true
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
