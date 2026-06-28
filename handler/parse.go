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

var mtx sync.Mutex

// Представляет структуру входящего HTTP-запроса с формулой
type ParserRequest struct {
	Formula string `json:"formula"`
}

// Ответ сервера содержит AST, результат проверки на существование контрпримера
// и граф Крипке
type ParserResponse struct {
	Success             bool                `json:"success"`
	Result              string              `json:"result"`
	Error               string              `json:"error"`
	ContralExampleCheck bool                `json:"conterexample check"`
	KripkeGraph         *solver.KripkeFrame `json:"Kripke graph"`
}

// ParseHandler обрабатывает POST-запросы для лексического анализа,
// парсинга формулы и поиска контрпримера в моделях Крипке
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

	// Инициализация библиотеки формул и упаковка её в контекст
	bibliothek := solver.NewFormulaBibliothek()
	ctx := solver.PackBibliothek(r.Context(), bibliothek)

	// Блокируем доступ к парсеру, так как текущая реализация
	// синтаксического анализатора хранит внутренее состояние и
	// не гарантирует потокобезопасность
	mtx.Lock()
	tokens := syntax.Lex(parserRequest.Formula)
	parser := syntax.NewParser(tokens)
	ast := parser.ParseExpression()
	defer mtx.Unlock()

	//Регистрируем формулу в библиотеке
	rootFormulaNumber := bibliothek.Bibliothek(ast)

	// Инициализируем контрмодели и создаем корневой мир
	conterModel := conterexample.NewContermodel()
	rootNumber := conterModel.NextWorldNumber()
	rootWorld := conterexample.NewModelWorld(rootNumber)

	// Пытаемся опровергнуть формулу
	//Примечание: это стандартный способ поиска контрпримеров в логике доказуемости
	rootWorld.FalseFormula = append(rootWorld.FalseFormula, rootFormulaNumber)
	conterModel.Frame.Worlds[rootNumber] = rootWorld

	//проверка общезначимости
	conterexampleCheck := conterModel.Prove(ctx, rootNumber, nil)

	parserResponse := ParserResponse{
		Success:             true,
		Result:              ast.String(),
		ContralExampleCheck: conterexampleCheck,
		KripkeGraph:         conterModel.InputToKripke(),
	}

	b, err := json.Marshal(parserResponse)
	if err != nil {
		fmt.Println("Ошибка при сериализации ответа:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		fmt.Println("Ошибка при отправке ответа:", err)
		return
	}
}
