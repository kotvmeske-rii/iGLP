package syntax

//Парсер - собираем дерево AddNote из токенов

type Parser struct {
	tokens []Token
	pos    int //pos??
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

//??????????????? токен на этой позиции??
func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokEOF}
	}
	return p.tokens[p.pos]
}

//???????? позиция??
func (p *Parser) consume() {
	p.pos++
}

//1. Implication
func (p *Parser) ParseExpression() Node {
	left := p.parseOr()
	if p.current().Type == TokImpl {
		p.consume()
		right := p.ParseExpression() //right assotoation
		return &BinaryNode{
			Op:    "->",
			Left:  left,
			Right: right,
		}
	}
	return left
}

//2. Or
func (p *Parser) parseOr() Node {
	left := p.parseAnd()
	if p.current().Type == TokOr {
		p.consume()
		right := p.parseAnd()
		return &BinaryNode{
			Op:    "|",
			Left:  left,
			Right: right,
		}
	}
	return left
}

//3. And
func (p *Parser) parseAnd() Node {
	left := p.parseUnary()
	if p.current().Type == TokAnd {
		p.consume()
		right := p.parseUnary()
		return &BinaryNode{
			Op:    "&",
			Left:  left,
			Right: right,
		}
	}
	return left
}

//4. Not and Box
func (p *Parser) parseUnary() Node {
	tok := p.current()

	if tok.Type == TokNot {
		p.consume()
		child := p.parseUnary() //реккурсивно парсим тч после ~
		return &UnaryNode{
			Op:    "~",
			Child: child,
		}
	}

	if tok.Type == TokBox {
		p.consume()
		child := p.parseUnary()
		return &UnaryNode{
			Op:    "[]",
			Child: child,
		}
	}
	return p.parsePrimary()
}

//5. Base
func (p *Parser) parsePrimary() Node {
	tok := p.current()

	if tok.Type == TokVar {
		p.consume()
		return &VarNode{
			Name: tok.Value,
		}
	}

	if tok.Type == TokLParent {
		p.consume()                 //
		expr := p.ParseExpression() //
		p.consume()                 //
		return expr
	}

	if tok.Type == TokFalse {
		p.consume()
		return &VarConst{
			Op: "_",
		}
	}

	if tok.Type == TokTrue {
		p.consume()
		return &VarConst{
			Op: "+",
		}
	}
	return nil
}

//функции вызывают друг друга по цепочке - соблюдаем приорететы операций
