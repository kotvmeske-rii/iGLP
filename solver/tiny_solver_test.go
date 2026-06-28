package solver

// import (
// 	"iglp/syntax"
// 	"testing"
// )

// func TestSolver(t *testing.T) {
// 	tests := []struct {
// 		name  string
// 		input string
// 		want  bool
// 	}{
// 		{
// 			name:  "denial",
// 			input: "p & ~q",
// 			want:  false,
// 		},
// 		{
// 			name:  "lie",
// 			input: "[](p & ~p)",
// 			want:  true,
// 		},
// 		{
// 			name:  "not box",
// 			input: "~[]~p",
// 			want:  false,
// 		},
// 		{
// 			name:  "box not",
// 			input: "[]~p",
// 			want:  true,
// 		},
// 		{
// 			name:  "Loeb axiom",
// 			input: "[]([]q -> q) -> []q",
// 			want:  true,
// 		},
// 	}

// 	frame := &KripkeFrame{
// 		Worlds: map[int]*ModelWorld{
// 			0: {Valuation: map[string]bool{"p": true, "q": false}},
// 			1: {Valuation: map[string]bool{"p": true, "q": true}},
// 			2: {Valuation: map[string]bool{"p": false, "q": false}},
// 		},
// 		Relations: map[int][]int{
// 			0: {1, 2},
// 			1: {2},
// 			2: {},
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Шаг А: Разбиваем строку на токены
// 			tokens := syntax.Lex(tc.input)

// 			// Шаг Б: Строим дерево
// 			parser := syntax.NewParser(tokens)
// 			ast := parser.ParseExpression()

// 			// Шаг В: Сверяем то, что получилось, с тем, что мы хотели (want)

// 			got := CheckFormula(ast, 2, frame)

// 			if got != tc.want {
// 				t.Errorf("\nОшибка в тесте '%s'\nВход:    %s\nПолучили: %t\nОжидали:  %t",
// 					tc.name, tc.input, got, tc.want)
// 			}
// 		})
// 	}
// }
