package chquery

func Parse(s string) (Tokens, error) {
	lex := newLexer(s)
	for {
		tok, err := lex.NextToken()
		if err != nil {
			return nil, err
		}
		if tok == eofToken {
			break
		}
	}
	return lex.tokens, nil
}
