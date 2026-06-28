package syntax

func Lex(input string) []Token {
	var tokens []Token
	runes := []rune(input) //[]rune??

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
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
			if i+1 < len(runes) && runes[i+1] == '>' {
				tokens = append(tokens, Token{Type: TokImpl})
				i++
			}
		case '_':
			tokens = append(tokens, Token{Type: TokFalse})
		case '+':
			tokens = append(tokens, Token{Type: TokTrue})
		case '[':
			if i+1 < len(runes) && runes[i+1] == ']' {
				tokens = append(tokens, Token{Type: TokBox})
				i++
			}
		default:
			if ch >= 'a' && ch <= 'z' {
				tokens = append(tokens, Token{Type: TokVar, Value: string(ch)})
			}
		}
	}

	return tokens
}
