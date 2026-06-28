package conterexample

import (
	"iglp/solver"
)

func NewModelWorld(id int) *solver.ModelWorld {
	return &solver.ModelWorld{
		Number:       id,
		Valuation:    make(map[solver.FormulaNumber]bool),
		TrueFormula:  []solver.FormulaNumber{},
		FalseFormula: []solver.FormulaNumber{},
	}
}

// clone for backtraking
func (c *Contermodel) cloneWorld(world *solver.ModelWorld) *solver.ModelWorld {
	newWorld := NewModelWorld(world.Number)
	newWorld.TrueFormula = append(newWorld.TrueFormula, world.TrueFormula...)
	newWorld.FalseFormula = append(newWorld.FalseFormula, world.FalseFormula...)

	for k, v := range world.Valuation {
		newWorld.Valuation[k] = v
	}

	return newWorld
}
