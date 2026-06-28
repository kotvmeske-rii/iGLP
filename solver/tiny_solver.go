package solver

import (
	"context"
	"iglp/syntax"
)

// ModelWorld описывает отдельный мир в модели Крипке
// Храним заведомо истиные и ложные формулы
// свой номер(ID) и означивание
type ModelWorld struct {
	Number       int
	Valuation    map[FormulaNumber]bool
	TrueFormula  []FormulaNumber
	FalseFormula []FormulaNumber
}

// Шкала Крипке - пара из миров
// и бинарного отношения достижимости между ними
type KripkeFrame struct {
	Worlds    map[int]*ModelWorld
	Relations map[int][]int
}

// Рекурсивно вычисляем истинность формулы в заданном мире
// true - формула истинна
func CheckFormula(ctx context.Context, number FormulaNumber, worldN int, kripkeFrame *KripkeFrame) bool {
	// Достаем библиотеку формул
	bibliothek := GetContext(ctx)
	key := bibliothek.Key(number)

	switch key.Op {
	case syntax.TokVar:
		return kripkeFrame.Worlds[worldN].Valuation[number]
	case syntax.TokFalse:
		// Константа "ложь" - ложна в любом мире
		return false
	case syntax.TokTrue:
		// Константа "истина" - истинна в любом мире
		return true
	case syntax.TokNot:
		// Истинно, если подформула ложна
		return !CheckFormula(ctx, key.Left, worldN, kripkeFrame)
	case syntax.TokBox:
		// []A истинна в мире w, если А истинна во всех мирах, достижимых из w
		// Примечание: помним, что должна сохраняться транзитивность
		achievWorld := kripkeFrame.Relations[worldN]

		for _, childWorldN := range achievWorld {
			if !CheckFormula(ctx, key.Left, childWorldN, kripkeFrame) {
				//нашли мир где подформула ложна
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
		//Примечание: для интуиционисткой логики определение будет иным
		//(истинность проверяется во всех достижимых мирах)
		return !CheckFormula(ctx, key.Left, worldN, kripkeFrame) ||
			CheckFormula(ctx, key.Right, worldN, kripkeFrame)
	}
	return false
}
