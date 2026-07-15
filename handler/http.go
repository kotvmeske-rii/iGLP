package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

// Представляет структуру входящего HTTP-запроса с формулой
type ParserRequest struct {
	Formula string `json:"formula"`
}

// Ответ сервера содержит AST, результат проверки на существование контрпримера
// и граф Крипке
type ParserResponse struct {
	Success       bool `json:"success"`
	*SolverResult `json:"data,omitempty"`
	Error         string `json:"error"`
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

	result, err := SolveResponse(r.Context(), parserRequest)

	if err != nil {
		log.Println("Ошибка при парсинге формулы:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// b, err := json.Marshal(parserResponse)
	// if err != nil {
	// 	fmt.Println("Ошибка при сериализации ответа:", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// if _, err := w.Write(b); err != nil {
	// 	fmt.Println("Ошибка при отправке ответа:", err)
	// 	return
	// }

	if err := json.NewEncoder(w).Encode(
		ParserResponse{
			Success:      true,
			SolverResult: result,
		},
	); err != nil {
		return
	}
}
