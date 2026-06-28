package syntax

//Лексический анализ входящей формулы
func Lex(input string) []Token {
	var tokens []Token
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		//Пропускаем пробелы
		if ch == ' ' {
			continue
		}

		switch ch {
		case '~':
			tokens = append(tokens, Token{Type: TokNot})
		case '&':
			tokens = append(tokens, Token{Type: TokAnd})
		case '|':
			tokens = append(tokens, Token{Type: TokOr})

		case '(':
			tokens = append(tokens, Token{Type: TokLParent})
		case ')':
			tokens = append(tokens, Token{Type: TokRParent})
		case '-':
			//Распознаем импликацию
			if i+1 < len(runes) && runes[i+1] == '>' {
				tokens = append(tokens, Token{Type: TokImpl})
				i++
			}
		case '_':
			tokens = append(tokens, Token{Type: TokFalse})
		case '+':
			tokens = append(tokens, Token{Type: TokTrue})
		case '[':
			//Распознаем box
			if i+1 < len(runes) && runes[i+1] == ']' {
				tokens = append(tokens, Token{Type: TokBox})
				i++
			}
		default:
			//Поддерживаются имена в виде строчных латинских букв
			if ch >= 'a' && ch <= 'z' {
				tokens = append(tokens, Token{Type: TokVar, Value: string(ch)})
			}
		}
	}

	return tokens
}
