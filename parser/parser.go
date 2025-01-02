package parser

import (
	"fmt"

	"github.com/gqlhub/gqlhub-core/ast"
	"github.com/gqlhub/gqlhub-core/lexer"
	"github.com/gqlhub/gqlhub-core/token"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) (*Parser, error) {
	p := &Parser{l: l}
	if err := p.next(); err != nil {
		return nil, fmt.Errorf("failed to initialize parser tokens: %w", err)
	}
	if err := p.next(); err != nil {
		return nil, fmt.Errorf("failed to initialize parser tokens: %w", err)
	}
	return p, nil
}

func (p *Parser) ParseDocument() (*ast.Document, error) {
	doc := &ast.Document{}

	for p.curToken.Type != token.EOF {
		def, err := p.parseDefinition()
		if err != nil {
			return nil, err
		}
		if def != nil {
			doc.Definitions = append(doc.Definitions, def)
		}

		if err = p.next(); err != nil {
			return nil, err
		}
	}

	return doc, nil
}

func (p *Parser) next() (err error) {
	p.curToken = p.peekToken
	p.peekToken, err = p.l.NextToken()
	return
}

func (p *Parser) expect(tokType token.Type) error {
	if p.curToken.Type != tokType {
		return fmt.Errorf("expected %s, got %s", tokType, p.curToken.Type)
	}
	return nil
}

func (p *Parser) parseDefinition() (ast.Definition, error) {
	if p.peekToken.Type == token.LBRACE {
		if err := p.next(); err != nil {
			return nil, err
		}
		return p.parseAnonymousOperation()
	}

	if p.peekToken.Type == token.NAME {
		switch p.curToken.Literal {
		case "query", "mutation", "subscription":
			return p.parseOperationDefinition()
		}
	}

	return nil, nil
}

func (p *Parser) parseOperationType() (ast.OperationType, error) {
	op := ast.OperationType(p.curToken.Literal)
	if err := p.next(); err != nil {
		return op, err
	}
	return op, nil
}

func (p *Parser) parseOperationDefinition() (*ast.OperationDefinition, error) {
	ot, err := p.parseOperationType()
	if err != nil {
		return nil, err
	}

	op := &ast.OperationDefinition{
		OperationType: ot,
	}

	if p.curToken.Type == token.NAME {
		op.Name = &ast.Name{
			Value: p.curToken.Literal,
		}
		if err := p.next(); err != nil {
			return nil, err
		}
	}

	if p.curToken.Type == token.LPAREN {
		vars, err := p.parseVariableDefinitions()
		if err != nil {
			return nil, err
		}
		op.VariableDefs = vars
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	op.Directives = directives

	ss, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	op.SelectionSet = ss

	return op, nil
}

func (p *Parser) parseVariableDefinitions() ([]*ast.VariableDefinition, error) {
	var defs []*ast.VariableDefinition

	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		def, err := p.parseVariableDefinition()
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}

	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	return defs, nil
}

func (p *Parser) parseVariableDefinition() (*ast.VariableDefinition, error) {
	vd := &ast.VariableDefinition{
		Position: p.curToken.Start,
	}

	val, err := p.parseVariable()
	if err != nil {
		return nil, err
	}
	variable, ok := val.(*ast.Variable)
	if !ok {
		return nil, fmt.Errorf("expected *ast.Variable, got %T", val)
	}
	vd.Variable = variable

	if err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	vd.Type = typ

	if p.curToken.Type == token.EQUALS {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		vd.DefaultValue = val
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	vd.Directives = directives

	return vd, nil
}

func (p *Parser) parseType() (ast.Type, error) {
	var typ ast.Type

	if p.curToken.Type == token.LBRACK {
		pos := p.curToken.Start

		if err := p.next(); err != nil {
			return nil, err
		}

		innerType, err := p.parseType()
		if err != nil {
			return nil, err
		}

		if err := p.expect(token.RBRACK); err != nil {
			return nil, err
		}
		if err := p.next(); err != nil {
			return nil, err
		}

		typ = &ast.ListType{
			Position: pos,
			Type:     innerType,
		}
	} else if p.curToken.Type == token.NAME {
		typ = &ast.NamedType{
			Position: p.curToken.Start,
			Name: &ast.Name{
				Position: p.curToken.Start,
				Value:    p.curToken.Literal,
			},
		}
		if err := p.next(); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unexpected token in type: %s", p.curToken.Type)
	}

	if p.curToken.Type == token.BANG {
		typ = &ast.NonNullType{
			Position: p.curToken.Start,
			Type:     typ,
		}
		if err := p.next(); err != nil {
			return nil, err
		}
	}

	return typ, nil
}

func (p *Parser) parseAnonymousOperation() (*ast.OperationDefinition, error) {
	if err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}

	op := &ast.OperationDefinition{
		OperationType: ast.OperationTypeQuery,
	}
	ss, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	op.SelectionSet = ss
	return op, nil
}

func (p *Parser) parseFragmentDefinition() (*ast.FragmentDefinition, error) {
	pos := p.curToken.Start

	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	frag := &ast.FragmentDefinition{
		Position: pos,
		Name: &ast.Name{
			Value: p.curToken.Literal,
		},
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	if p.curToken.Literal != "on" {
		return nil, fmt.Errorf("expected 'on', got %s", p.curToken.Literal)
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	frag.TypeCondition = &ast.NamedType{
		Name: &ast.Name{
			Value: p.curToken.Literal,
		},
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	ss, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	frag.SelectionSet = ss

	return frag, nil
}

func (p *Parser) parseSelectionSet() (*ast.SelectionSet, error) {
	pos := p.curToken.Start

	if err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	ss := &ast.SelectionSet{
		Position: pos,
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		sel, err := p.parseSelection()
		if err != nil {
			return nil, err
		}
		ss.Selections = append(ss.Selections, sel)
	}

	if err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	return ss, nil
}

func (p *Parser) parseSelection() (ast.Selection, error) {
	if p.curToken.Type == token.SPREAD {
		return p.parseFragment()
	}
	return p.parseField()
}

func (p *Parser) parseField() (ast.Selection, error) {
	field := &ast.Field{
		Position: p.curToken.Start,
	}

	if p.curToken.Type == token.NAME && p.peekToken.Type == token.COLON {
		field.Alias = &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		if err := p.expect(token.COLON); err != nil {
			return nil, err
		}
		if err := p.next(); err != nil {
			return nil, err
		}
	}

	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	field.Name = &ast.Name{
		Position: p.curToken.Start,
		Value:    p.curToken.Literal,
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArguments()
		if err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}
		field.Arguments = args
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	field.Directives = directives

	if p.curToken.Type == token.LBRACE {
		ss, err := p.parseSelectionSet()
		if err != nil {
			return nil, err
		}
		field.SelectionSet = ss
	}

	return field, nil
}

func (p *Parser) parseFragment() (ast.Selection, error) {
	if err := p.next(); err != nil {
		return nil, err
	}
	if p.curToken.Literal == "on" {
		return p.parseInlineFragment()
	}
	return p.parseFragmentSpread()
}

func (p *Parser) parseInlineFragment() (*ast.InlineFragment, error) {
	inf := &ast.InlineFragment{
		Position: p.curToken.Start,
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	nt, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	inf.TypeCondition = nt

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	inf.Directives = directives

	ss, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	inf.SelectionSet = ss

	return inf, nil
}

func (p *Parser) parseNamedType() (nt *ast.NamedType, err error) {
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	nt = &ast.NamedType{
		Position: p.curToken.Start,
		Name: &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		},
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	return nt, nil
}

func (p *Parser) parseDirectives() ([]*ast.Directive, error) {
	var directives []*ast.Directive

	for p.curToken.Type == token.AT {
		directive, err := p.parseDirective()
		if err != nil {
			return directives, err
		}
		directives = append(directives, directive)
	}

	return directives, nil
}

func (p *Parser) parseDirective() (*ast.Directive, error) {
	pos := p.curToken.Start

	if err := p.expect(token.AT); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	dir := &ast.Directive{
		Position: pos,
		Name: &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		},
	}
	if err := p.next(); err != nil {
		return dir, err
	}

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArguments()
		if err != nil {
			return dir, err
		}
		dir.Arguments = args
	}

	return dir, nil
}

func (p *Parser) parseArguments() ([]*ast.Argument, error) {
	var args []*ast.Argument

	if err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		arg, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	if err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	return args, nil
}

func (p *Parser) parseArgument() (*ast.Argument, error) {
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	arg := &ast.Argument{
		Position: p.curToken.Start,
		Name: &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		},
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	if err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	arg.Value = value

	return arg, nil
}

func (p *Parser) parseValue() (ast.Value, error) {
	switch p.curToken.Type {
	case token.INT:
		val := &ast.IntValue{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		return val, nil
	case token.FLOAT:
		val := &ast.FloatValue{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		return val, nil
	case token.STRING_VALUE:
		val := &ast.StringValue{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
			Block:    false,
		}
		if err := p.next(); err != nil {
			return nil, err
		}
		return val, nil
	//case token.STRING_VALUE:
	//	val := &ast.StringValue{Value: p.curToken.Literal, Block: true}
	//	if err := p.next(); err != nil {
	//		return val, err
	//	}
	//	return val, nil
	case token.NAME:
		switch p.curToken.Literal {
		case "true":
			val := &ast.BooleanValue{
				Position: p.curToken.Start,
				Value:    true,
			}
			if err := p.next(); err != nil {
				return nil, err
			}
			return val, nil
		case "false":
			val := &ast.BooleanValue{
				Position: p.curToken.Start,
				Value:    false,
			}
			if err := p.next(); err != nil {
				return nil, err
			}
			return val, nil
		case "null":
			val := &ast.NullValue{
				Position: p.curToken.Start,
			}
			if err := p.next(); err != nil {
				return nil, err
			}
			return val, nil
		default:
			val := &ast.EnumValue{
				Position: p.curToken.Start,
				Value:    p.curToken.Literal,
			}
			if err := p.next(); err != nil {
				return nil, err
			}
			return val, nil
		}
	case token.DOLLAR:
		return p.parseVariable()

	case token.LBRACK:
		return p.parseListValue()

	case token.LBRACE:
		return p.parseObjectValue()

	default:
		return nil, fmt.Errorf("unexpected value token: %s", p.curToken.Type)
	}
}

func (p *Parser) parseListValue() (ast.Value, error) {
	list := &ast.ListValue{
		Position: p.curToken.Start,
	}

	if err := p.expect(token.LBRACK); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACK && p.curToken.Type != token.EOF {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		list.Values = append(list.Values, val)
	}

	if err := p.expect(token.RBRACK); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	return list, nil
}

func (p *Parser) parseObjectValue() (ast.Value, error) {
	obj := &ast.ObjectValue{
		Position: p.curToken.Start,
	}

	if err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		field, err := p.parseObjectField()
		if err != nil {
			return nil, err
		}
		obj.Fields = append(obj.Fields, field)
	}

	if err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	return obj, nil
}

func (p *Parser) parseObjectField() (*ast.ObjectField, error) {
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	of := &ast.ObjectField{
		Position: p.curToken.Start,
		Name: &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		},
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	if err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	of.Value = val

	return of, nil
}

func (p *Parser) parseVariable() (ast.Value, error) {
	pos := p.curToken.Start

	if err := p.expect(token.DOLLAR); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}

	variable := &ast.Variable{
		Position: pos,
		Name: &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		},
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	return variable, nil
}

func (p *Parser) parseFragmentSpread() (*ast.FragmentSpread, error) {
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	fs := &ast.FragmentSpread{
		Position: p.curToken.Start,
		Name: &ast.Name{
			Position: p.curToken.Start,
			Value:    p.curToken.Literal,
		},
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	fs.Directives = directives

	return fs, nil
}
