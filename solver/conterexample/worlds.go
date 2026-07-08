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
