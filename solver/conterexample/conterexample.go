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
func (c *Contermodel) Prove(ctx context.Context, worldNumber int, history []solver.FormulaNumber) bool {
	world := c.Frame.Worlds[worldNumber]

	//Step 1: prove denial
	if c.Denial(world) {
		return false //невозможно чтобы формула в мире была одновременно и верна и не верна
	}

	bibliothek := solver.GetContext(ctx)

	//Step 2: local world

	//constants
	for _, v := range world.TrueFormula {
		if bibliothek.Key(v).Op == syntax.TokFalse {
			return false
		}
	}

	for _, v := range world.FalseFormula {
		if bibliothek.Key(v).Op == syntax.TokTrue {
			return false
		}
	}

	for k, v := range world.TrueFormula {
		if bibliothek.Key(v).Op == syntax.TokTrue {
			world.TrueFormula = append(world.TrueFormula[:k], world.TrueFormula[k+1:]...)
			return c.Prove(ctx, worldNumber, history)
		}
	}

	for k, v := range world.FalseFormula {
		if bibliothek.Key(v).Op == syntax.TokFalse {
			world.FalseFormula = append(world.FalseFormula[:k], world.FalseFormula[k+1:]...)
			return c.Prove(ctx, worldNumber, history)
		}
	}

	//true impl
	for k, v := range world.TrueFormula {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokImpl {
			backup := c.cloneWorld(world)
			leftKey := key.Left
			rightKey := key.Right

			//delate impl
			world.TrueFormula = append(world.TrueFormula[:k], world.TrueFormula[k+1:]...)

			//left false
			world.FalseFormula = append(world.FalseFormula, leftKey)
			if c.Prove(ctx, worldNumber, history) {
				return true
			}

			//right true
			c.Frame.Worlds[worldNumber] = backup
			world = c.Frame.Worlds[worldNumber]
			world.TrueFormula = append(world.TrueFormula[:k], world.TrueFormula[k+1:]...)
			world.TrueFormula = append(world.TrueFormula, rightKey)
			return c.Prove(ctx, worldNumber, history)
		}
	}

	//false impl
	for k, v := range world.FalseFormula {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokImpl {
			leftKey := key.Left
			rightKey := key.Right

			world.FalseFormula = append(world.FalseFormula[:k], world.FalseFormula[k+1:]...)
			world.TrueFormula = append(world.TrueFormula, leftKey)
			world.FalseFormula = append(world.FalseFormula, rightKey)

			return c.Prove(ctx, worldNumber, history)
		}
	}

	//variate true
	for k, v := range world.TrueFormula {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokVar {
			world.Valuation[v] = true
			world.TrueFormula = append(world.TrueFormula[:k], world.TrueFormula[k+1:]...)
			return c.Prove(ctx, worldNumber, history)
		}
	}

	//var false
	for k, v := range world.FalseFormula {
		key := bibliothek.Key(v)
		if key.Op == syntax.TokVar {
			world.Valuation[v] = false
			world.FalseFormula = append(world.FalseFormula[:k], world.FalseFormula[k+1:]...)
			return c.Prove(ctx, worldNumber, history)
		}
	}

	//Step 3: next world
	//false box
	for k, v := range world.FalseFormula {

		key := bibliothek.Key(v)

		if key.Op == syntax.TokBox {
			//иначе мы сломаемся на чём-то вроде формулы Лёба []([]p -> p) -> []p
			//из-за нехватки памяти, тк цепочка не будет нётеровой
			looped := false
			for _, h := range history {
				if h == v {
					looped = true
					break
				}
			}

			if looped {
				return false
			}

			childNumber := key.Left

			//delet box
			world.FalseFormula = append(world.FalseFormula[:k], world.FalseFormula[k+1:]...)

			newNumber := c.NextWorldNumber()
			newWorld := NewModelWorld(newNumber)
			newWorld.FalseFormula = append(newWorld.FalseFormula, childNumber, v)

			//if in w_0 []B - true => in w_1 B, []B - true
			for _, v := range world.TrueFormula {
				trueKey := bibliothek.Key(v)

				if trueKey.Op == syntax.TokBox {
					newWorld.TrueFormula = append(newWorld.TrueFormula, trueKey.Left, v)
				}
			}

			c.Frame.Worlds[newNumber] = newWorld
			c.Frame.Relations[worldNumber] = append(c.Frame.Relations[worldNumber], newNumber)

			newHistory := append(history, v)

			if c.Prove(ctx, newNumber, newHistory) {
				return true
			}

			return false
		}
	}

	return true
}

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
