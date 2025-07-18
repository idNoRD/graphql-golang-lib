package parser

import "github.com/gqlhub/gqlhub-core/token"

func IsStringValue(tok token.Type) bool {
	return tok == token.STRING || tok == token.BLOCK_STRING
}
