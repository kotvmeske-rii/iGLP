package syntax

type TokenType int

const (
	TokEOF TokenType = iota
	TokVar
	TokNot     // ~
	TokBox     // []
	TokAnd     // &
	TokOr      // |
	TokImpl    // ->
	TokLParent // (
	TokRParent // )
	TokFalse   // _
	TokTrue    // +
)

type Token struct {
	Type  TokenType
	Value string
}
