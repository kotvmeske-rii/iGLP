package solver

import (
	"context"
	"iglp/syntax"
)

type ModelWorld struct {
	Number       int
	Valuation    map[FormulaNumber]bool
	TrueFormula  []FormulaNumber
	FalseFormula []FormulaNumber
}

type KripkeFrame struct {
	Worlds    map[int]*ModelWorld
	Relations map[int][]int
}

func CheckFormula(ctx context.Context, number FormulaNumber, worldN int, kripkeFrame *KripkeFrame) bool {
	bibliothek := GetContext(ctx)
	key := bibliothek.Key(number)

	switch key.Op {
	case syntax.TokVar:
		return kripkeFrame.Worlds[worldN].Valuation[number]
	case syntax.TokFalse:
		return false
	case syntax.TokTrue:
		return true
	case syntax.TokNot:
		return !CheckFormula(ctx, key.Left, worldN, kripkeFrame)
	case syntax.TokBox:
		achievWorld := kripkeFrame.Relations[worldN]

		for _, childWorldN := range achievWorld {
			if !CheckFormula(ctx, key.Left, childWorldN, kripkeFrame) {
				return false
			}
		}
		return true
	case syntax.TokAnd:
		return CheckFormula(ctx, key.Left, worldN, kripkeFrame) &&
			CheckFormula(ctx, key.Right, worldN, kripkeFrame)
	case syntax.TokOr:
		return CheckFormula(ctx, key.Left, worldN, kripkeFrame) ||
			CheckFormula(ctx, key.Right, worldN, kripkeFrame)
	case syntax.TokImpl:
		return !CheckFormula(ctx, key.Left, worldN, kripkeFrame) ||
			CheckFormula(ctx, key.Right, worldN, kripkeFrame)
	}
	return false
}
