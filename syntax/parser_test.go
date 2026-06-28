package syntax

import "testing"

func TestParseExpression(t *testing.T) {
	tests := []struct {
		name  string // Название теста
		input string // Что даем на вход парсеру
		want  string // Какое дерево ожидаем получить на выходе
	}{
		{
			name:  "Простая переменная",
			input: "p",
			want:  "p",
		},
		{
			name:  "Приоритет унарного отрицания",
			input: "p & ~q",
			want:  "(p & ~(q))", // Отрицание должно примениться только к p
		},
		{
			name:  "Приоритет конъюнкции над импликацией",
			input: "p & q -> r",
			want:  "((p & q) -> r)", // Сначала И, потом стрелочка
		},
		{
			name:  "Аксиома К модальной логики",
			input: "[](p -> q) -> ([]p -> []q)",
			want:  "([]((p -> q)) -> ([](p) -> [](q)))",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Шаг А: Разбиваем строку на токены
			tokens := Lex(tc.input)

			// Шаг Б: Строим дерево
			parser := &Parser{tokens: tokens, pos: 0}
			ast := parser.ParseExpression()

			// Шаг В: Сверяем то, что получилось, с тем, что мы хотели (want)
			got := ast.String()

			if got != tc.want {
				t.Errorf("\nОшибка в тесте '%s'\nВход:    %s\nПолучили: %s\nОжидали:  %s",
					tc.name, tc.input, got, tc.want)
			}
		})
	}
}
