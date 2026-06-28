package handler

import (
	"encoding/json"
	"fmt"
	"iglp/solver"
	"iglp/solver/conterexample"
	"iglp/syntax"
	"net/http"
	"sync"
)

// mtx защищает глобальное состояние парсера
// Примечание: учитывая потокобезопасность синтаксического анализатора
// в пакете syntax, стоит пересмотреть необходимость глобальной блокировки
// в пользу локальных экземпляров для повышения пропускной способности
var mtx sync.Mutex

type ParserRequest struct {
	Formula string `json:"formula"`
}

type ParserResponse struct {
	Success             bool                `json:"success"`
	Result              string              `json:"result"`
	Error               string              `json:"error"` //если ошибка будет "" не пустым
	ContralExampleCheck bool                `json:"conterexample check"`
	KripkeGraph         *solver.KripkeFrame `json:"Kripke graph"`
}

func ParseHandler(w http.ResponseWriter, r *http.Request) {
	var parserRequest ParserRequest

	// Декодируем входящий JSON
	// Ограничиваем тело запроса для предотвращения атак, связанных с исчерпанием памяти,
	// если формула окажется аномально большой
	if err := json.NewDecoder(r.Body).Decode(&parserRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(
			ParserResponse{
				Success: false,
				Error:   "Некорректный формат JSON или синтаксическая ошибка в теле запроса",
			},
		); err != nil {
			return
		}
		return
	}

	bibliothek := solver.NewFormulaBibliothek()
	ctx := solver.PackBibliothek(r.Context(), bibliothek)
	// Блокируем доступ к парсеру, так как текущая реализация
	// синтаксического анализатора не гарантирует потокобезопасность.
	mtx.Lock()
	defer mtx.Unlock()

	tokens := syntax.Lex(parserRequest.Formula)
	parser := syntax.NewParser(tokens)
	// Построение AST. В случае некорректной формулы парсер
	// должен выбрасывать кастомную ошибку, которую мы обработаем отдельно
	ast := parser.ParseExpression()

	rootFormulaNumber := bibliothek.Bibliothek(ast)

	conterModel := conterexample.NewContermodel()
	rootNumber := conterModel.NextWorldNumber()
	rootWorld := conterexample.NewModelWorld(rootNumber)

	//we want refute formula, so we put they in FalseFormula in rootWorld
	rootWorld.FalseFormula = append(rootWorld.FalseFormula, rootFormulaNumber)
	conterModel.Frame.Worlds[rootNumber] = rootWorld

	conterexampleCheck := conterModel.Prove(ctx, rootNumber, nil)

	parserResponse := ParserResponse{
		Success:             true,
		Result:              ast.String(),
		ContralExampleCheck: conterexampleCheck,
		KripkeGraph:         conterModel.InputToKripke(),
	}

	b, err := json.Marshal(parserResponse)
	if err != nil {
		// Логируем ошибку на серверную сторону, пользователю возвращаем 500
		fmt.Println("Ошибка при сериализации ответа:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		fmt.Println("Ошибка при отправке ответа:", err)
		return
	}
}
