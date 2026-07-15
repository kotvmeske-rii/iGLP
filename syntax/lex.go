package syntax

import "fmt"

type Lexer struct {
	position int
	chars    []rune
}

func (l *Lexer) EOF() bool {
	return l.position >= len(l.chars)
}

func (l *Lexer) currentChar() rune {
	if l.EOF() {
		return 0
	}

	return l.chars[l.position]
}

func (l *Lexer) nextChar() rune {
	if l.position+1 >= len(l.chars) {
		return 0
	}

	return l.chars[l.position+1]
}

func (l *Lexer) nextPosition() {
	l.position++
}

// Лексический анализ входящей формулы
func Lex(input string) []Token {
	var tokens []Token
	l := &Lexer{
		position: 0,
		chars:    []rune(input),
	}

	for !l.EOF() {
		ch := l.currentChar()

		//Пропускаем пробелы
		if ch == ' ' {
			l.nextPosition()

			continue
		}

		switch ch {
		case '~':
			tokens = append(
				tokens,
				Token{
					Type: TokNot,
				},
			)

			l.nextPosition()
		case '&':
			tokens = append(
				tokens,
				Token{
					Type: TokAnd,
				},
			)

			l.nextPosition()
		case '|':
			tokens = append(
				tokens, Token{
					Type: TokOr,
				},
			)

			l.nextPosition()
		case '(':
			tokens = append(
				tokens,
				Token{
					Type: TokLParent,
				},
			)

			l.nextPosition()
		case ')':
			tokens = append(
				tokens,
				Token{
					Type: TokRParent,
				},
			)

			l.nextPosition()
		case '_':
			tokens = append(
				tokens,
				Token{
					Type: TokFalse,
				},
			)

			l.nextPosition()
		case '+':
			tokens = append(
				tokens,
				Token{
					Type: TokTrue,
				},
			)

			l.nextPosition()
		case '-':
			//Распознаем импликацию
			if l.nextChar() == '>' {
				tokens = append(
					tokens,
					Token{
						Type: TokImpl,
					},
				)

				l.nextPosition()
				l.nextPosition()
			} else {
				fmt.Println("Undefined symbol")
				l.nextPosition()
			}
		case '[':
			//Распознаем box
			if l.nextChar() == ']' {
				tokens = append(tokens,
					Token{
						Type: TokBox,
					},
				)

				l.nextPosition()
				l.nextPosition()
			} else {
				fmt.Println("Undefined symbol")

				l.nextPosition()
			}
		default:
			//Поддерживаются имена в виде строчных латинских букв
			if letter(ch) {
				tokens = append(tokens, l.readLetter())

				l.nextPosition()
			}
		}

	}

	return tokens
}

func letter(ch rune) bool {
	return ch >= 'a' && ch <= 'z'
}

func (l *Lexer) readLetter() Token {
	begin := l.position

	if !l.EOF() && letter(l.currentChar()) {
		l.nextPosition()
	}

	return Token{
		Type:  TokVar,
		Value: string(l.chars[begin:l.position]),
	}
}
