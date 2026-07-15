package handler

import (
	"context"
	"iglp/conterexample"
	"iglp/solver"
	"iglp/syntax"
	"sync"
)

var mtx sync.Mutex

type SolverResult struct {
	Result              string
	ContralExampleCheck bool
	KripkeGraph         *solver.KripkeFrame
}

// Инициализация библиотеки формул и упаковка её в контекст
func ParseBibliothek(context context.Context) (context.Context, *solver.FormulaBibliothek) {
	bibliothek := solver.NewFormulaBibliothek()
	ctx := solver.PackBibliothek(context, bibliothek)

	return ctx, bibliothek
}

// парсим формулу
func ParseWork(parserRequest ParserRequest) syntax.Node {
	mtx.Lock()
	tokens := syntax.Lex(parserRequest.Formula)
	parser := syntax.NewParser(tokens)
	ast := parser.ParseExpression()
	defer mtx.Unlock()

	return ast
}

// Регистрируем формулу в библиотеке
func ParseRegistryFormula(ast syntax.Node, bibliothek *solver.FormulaBibliothek) solver.FormulaNumber {
	rootFormulaNumber := bibliothek.Bibliothek(ast)

	return rootFormulaNumber
}

//Работа с контрмоделями

func ParseWorkWithContermodel(
	ctx context.Context,
	rootFormulaNumber solver.FormulaNumber,
) (
	bool,
	*conterexample.Contermodel,
) {
	// Инициализируем контрмодели и создаем корневой мир
	conterModel := conterexample.NewContermodel()
	rootNumber := conterModel.NextWorldNumber()
	rootWorld := conterexample.NewModelWorld(rootNumber)

	// Пытаемся опровергнуть формулу
	//Примечание: это стандартный способ поиска контрпримеров в логике доказуемости
	rootWorld.FalseFormula = append(rootWorld.FalseFormula, rootFormulaNumber)
	conterModel.Frame.Worlds[rootNumber] = rootWorld

	//проверка общезначимости
	trace, conterexampleCheck := conterModel.Prove(ctx, rootNumber, 0, 0)

	if conterexampleCheck && trace != nil {
		conterModel.Sammeln(trace)
	}

	return conterexampleCheck, conterModel
}

func SolveResponse(context context.Context, parserRequest ParserRequest) (*SolverResult, error) {
	ctx, bibliothek := ParseBibliothek(context)
	ast := ParseWork(parserRequest)
	rootFormulaNumber := ParseRegistryFormula(ast, bibliothek)
	conterexampleCheck, conterModel := ParseWorkWithContermodel(ctx, rootFormulaNumber)

	return &SolverResult{
		Result:              ast.String(),
		ContralExampleCheck: conterexampleCheck,
		KripkeGraph:         conterModel.InputToKripke(),
	}, nil
}
