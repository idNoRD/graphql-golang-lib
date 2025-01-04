package parser

import "github.com/gqlhub/gqlhub-core/token"

func isDescription(tok token.Type) bool {
	return tok == token.STRING || tok == token.BLOCK_STRING
}
