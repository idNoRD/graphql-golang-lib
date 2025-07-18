package lexer

import (
	"github.com/gqlhub/gqlhub-core/token"
	"testing"
)

func BenchmarkDecodeRunePerformance(b *testing.B) {
	input := `
query getUser($userId: ID = 100, $withName: Boolean!) { # get user by id
  user(id: $userId) {
    id
    name @include(if: $withName)
    ...UserFields
  }
}
`

	l := New(input)

	b.ResetTimer()
	for {
		tok, err := l.NextToken()
		if err != nil {
			b.Fatalf("Error: %v", err)
		}
		if tok.Type == token.EOF {
			break
		}
	}
}
