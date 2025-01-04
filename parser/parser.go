package parser

import (
	"fmt"
	"strings"

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

func (p *Parser) next() error {
	p.curToken = p.peekToken
	var err error
	p.peekToken, err = p.l.NextToken()
	return err
}

func (p *Parser) expect(expectedToken token.Type) error {
	if p.curToken.Type != expectedToken {
		return fmt.Errorf("expected %s, got %s", expectedToken, p.curToken.Type)
	}
	return nil
}

func (p *Parser) expectOneOf(expectedTokens ...token.Type) error {
	for _, expected := range expectedTokens {
		if p.curToken.Type == expected {
			return nil
		}
	}

	var expectedStrings []string
	for _, tok := range expectedTokens {
		expectedStrings = append(expectedStrings, tok.String())
	}

	return fmt.Errorf("expected one of [%s], got %s", strings.Join(expectedStrings, ", "), p.curToken.Type)
}

func (p *Parser) expectAndAdvance(expectedToken token.Type) error {
	if p.curToken.Type != expectedToken {
		return fmt.Errorf("expected %s, got %s", expectedToken, p.curToken.Type)
	}
	return p.next()
}

func (p *Parser) expectLiteralAndAdvance(lit string) error {
	if p.curToken.Literal != lit {
		return fmt.Errorf("expected %s, got %s", lit, p.curToken.Literal)
	}
	return p.next()
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
		case "schema":
			return p.parseSchemaDefinition()
		case "scalar":
			return p.parseScalarTypeDefinition()
		case "type":
			return p.parseObjectTypeDefinition()
		case "interface":
			return p.parseInterfaceTypeDefinition()
		case "union":
			return p.parseUnionTypeDefinition()
		case "enum":
			return p.parseEnumTypeDefinition()
		case "input":
			return p.parseInputObjectTypeDefinition()
		case "directive":
			return p.parseDirectiveDefinition()
		case "query", "mutation", "subscription":
			return p.parseOperationDefinition()
		case "fragment":
			return p.parseFragmentDefinition()
		case "extend":
			return p.parseTypeSystemExtension()
		}
	}

	return nil, fmt.Errorf("unexpected keyword %s", p.curToken.Literal)
}

func (p *Parser) parseSchemaDefinition() (ast.Definition, error) {
	sd := &ast.SchemaDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		sd.Description = desc
	}

	if err := p.expectLiteralAndAdvance("schema"); err != nil {
		return nil, err
	}

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	sd.Directives = dirs

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	var roots []*ast.RootOperationTypeDefinition
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		rootDef, err := p.parseRootOperationTypeDefinition()
		if err != nil {
			return nil, err
		}
		roots = append(roots, rootDef)
	}
	sd.RootOperationDefs = roots

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return sd, nil
}

func (p *Parser) parseScalarTypeDefinition() (ast.Definition, error) {
	sdef := &ast.ScalarTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		sdef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("scalar"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	sdef.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	sdef.Directives = dirs

	return sdef, nil
}

func (p *Parser) parseObjectTypeDefinition() (ast.Definition, error) {
	odef := &ast.ObjectTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		odef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("type"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	odef.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		odef.Interfaces = ii
	}

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	odef.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		odef.Fields = fields
	}

	return odef, nil
}

func (p *Parser) parseInterfaceTypeDefinition() (ast.Definition, error) {
	idef := &ast.InterfaceTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		idef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("interface"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	idef.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		idef.Interfaces = ii
	}

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	idef.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		idef.Fields = fields
	}

	return idef, nil
}

func (p *Parser) parseUnionTypeDefinition() (ast.Definition, error) {
	udef := &ast.UnionTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		udef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("union"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	udef.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	udef.Directives = dirs

	if p.curToken.Type == token.EQUALS {
		types, err := p.parseUnionMemberTypes()
		if err != nil {
			return nil, err
		}
		udef.Types = types
	}

	return udef, nil
}

func (p *Parser) parseEnumTypeDefinition() (ast.Definition, error) {
	edef := &ast.EnumTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		edef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("enum"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	edef.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	edef.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		vals, err := p.parseEnumValuesDefinition()
		if err != nil {
			return nil, err
		}
		edef.Values = vals
	}

	return edef, nil
}

func (p *Parser) parseInputObjectTypeDefinition() (ast.Definition, error) {
	iod := &ast.InputObjectTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		iod.Description = desc
	}

	if err := p.expectLiteralAndAdvance("input"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	iod.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	iod.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseInputFieldsDefinition()
		if err != nil {
			return nil, err
		}
		iod.Fields = fields
	}

	return iod, nil
}

func (p *Parser) parseDirectiveDefinition() (ast.Definition, error) {
	ddef := &ast.DirectiveDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		ddef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("directive"); err != nil {
		return nil, err
	}
	if err := p.expectAndAdvance(token.AT); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ddef.Name = name

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArgumentsDefinition()
		if err != nil {
			return nil, err
		}
		ddef.Arguments = args
	}

	if p.curToken.Literal == "repeatable" {
		if err := p.next(); err != nil {
			return nil, err
		}
		ddef.Repeatable = true
	}

	if err := p.expectLiteralAndAdvance("on"); err != nil {
		return nil, err
	}

	locs, err := p.parseDirectiveLocations()
	if err != nil {
		return nil, err
	}
	ddef.Locations = locs

	return ddef, nil
}

func (p *Parser) parseDirectiveLocations() (ast.DirectiveLocations, error) {
	var types ast.DirectiveLocations
	for {
		nt, err := p.parseName()
		if err != nil {
			return nil, err
		}
		types = append(types, nt)
		if p.curToken.Type == token.PIPE {
			if err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}
	return types, nil
}

func (p *Parser) parseOperationType() (ast.OperationType, error) {
	if err := p.expect(token.NAME); err != nil {
		return "", err
	}
	opType := ast.OperationType(p.curToken.Literal)
	// TODO: fmt.Errorf("unknown root operation type: %q", p.curToken.Literal)
	if err := p.next(); err != nil {
		return "", err
	}
	return opType, nil
}

func (p *Parser) parseOperationDefinition() (*ast.OperationDefinition, error) {
	op := &ast.OperationDefinition{
		Position: p.curToken.Start,
	}

	opType, err := p.parseOperationType()
	if err != nil {
		return nil, err
	}
	op.OperationType = opType

	if p.curToken.Type == token.NAME {
		name, err := p.parseName()
		if err != nil {
			return nil, err
		}
		op.Name = name
	}

	if p.curToken.Type == token.LPAREN {
		varDefs, err := p.parseVariableDefinitions()
		if err != nil {
			return nil, err
		}
		op.VariableDefs = varDefs
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	op.Directives = directives

	selectionSet, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	op.SelectionSet = selectionSet

	return op, nil
}

func (p *Parser) parseName() (*ast.Name, error) {
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	name := &ast.Name{
		Position: p.curToken.Start,
		Value:    p.curToken.Literal,
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	return name, nil
}

func (p *Parser) parseVariableDefinitions() ([]*ast.VariableDefinition, error) {
	var defs []*ast.VariableDefinition

	if err := p.expectAndAdvance(token.LPAREN); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		def, err := p.parseVariableDefinition()
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}

	if err := p.expectAndAdvance(token.RPAREN); err != nil {
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

	if err := p.expectAndAdvance(token.COLON); err != nil {
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
	pos := p.curToken.Start

	if p.curToken.Type == token.LBRACK {
		if err := p.next(); err != nil {
			return nil, err
		}

		innerType, err := p.parseType()
		if err != nil {
			return nil, err
		}

		if err := p.expectAndAdvance(token.RBRACK); err != nil {
			return nil, err
		}

		typ = &ast.ListType{
			Position: pos,
			Type:     innerType,
		}
	} else if p.curToken.Type == token.NAME {
		var err error
		typ, err = p.parseNamedType()
		if err != nil {
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
	frag := &ast.FragmentDefinition{
		Position: p.curToken.Start,
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	frag.Name = name

	if p.curToken.Literal != "on" {
		return nil, fmt.Errorf("expected 'on', got %s", p.curToken.Literal)
	}
	if err := p.next(); err != nil {
		return nil, err
	}

	namedType, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	frag.TypeCondition = namedType

	ss, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	frag.SelectionSet = ss

	return frag, nil
}

func (p *Parser) parseSelectionSet() (*ast.SelectionSet, error) {
	ss := &ast.SelectionSet{
		Position: p.curToken.Start,
	}

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		sel, err := p.parseSelection()
		if err != nil {
			return nil, err
		}
		ss.Selections = append(ss.Selections, sel)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
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
		alias, err := p.parseName()
		if err != nil {
			return nil, err
		}
		field.Alias = alias

		if err := p.next(); err != nil {
			return nil, err
		}
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	field.Name = name

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArguments()
		if err != nil {
			return nil, err
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

func (p *Parser) parseNamedType() (*ast.NamedType, error) {
	nt := &ast.NamedType{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	nt.Name = name

	return nt, nil
}

func (p *Parser) parseDirectives() ([]*ast.Directive, error) {
	var directives []*ast.Directive

	for p.curToken.Type == token.AT {
		directive, err := p.parseDirective()
		if err != nil {
			return nil, err
		}
		directives = append(directives, directive)
	}

	return directives, nil
}

func (p *Parser) parseDirective() (*ast.Directive, error) {
	dir := &ast.Directive{
		Position: p.curToken.Start,
	}

	if err := p.expectAndAdvance(token.AT); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	dir.Name = name

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArguments()
		if err != nil {
			return nil, err
		}
		dir.Arguments = args
	}

	return dir, nil
}

func (p *Parser) parseArguments() ([]*ast.Argument, error) {
	var args []*ast.Argument

	if err := p.expectAndAdvance(token.LPAREN); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		arg, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	if err := p.expectAndAdvance(token.RPAREN); err != nil {
		return nil, err
	}

	return args, nil
}

func (p *Parser) parseArgument() (*ast.Argument, error) {
	arg := &ast.Argument{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	arg.Name = name

	if err := p.expectAndAdvance(token.COLON); err != nil {
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
	case token.STRING, token.BLOCK_STRING:
		return p.parseStringValue()
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

	if err := p.expectAndAdvance(token.LBRACK); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACK && p.curToken.Type != token.EOF {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		list.Values = append(list.Values, val)
	}

	if err := p.expectAndAdvance(token.RBRACK); err != nil {
		return nil, err
	}

	return list, nil
}

func (p *Parser) parseObjectValue() (ast.Value, error) {
	obj := &ast.ObjectValue{
		Position: p.curToken.Start,
	}

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		field, err := p.parseObjectField()
		if err != nil {
			return nil, err
		}
		obj.Fields = append(obj.Fields, field)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return obj, nil
}

func (p *Parser) parseObjectField() (*ast.ObjectField, error) {
	of := &ast.ObjectField{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	of.Name = name

	if err := p.expectAndAdvance(token.COLON); err != nil {
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
	variable := &ast.Variable{
		Position: p.curToken.Start,
	}

	if err := p.expectAndAdvance(token.DOLLAR); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	variable.Name = name

	return variable, nil
}

func (p *Parser) parseFragmentSpread() (*ast.FragmentSpread, error) {
	fs := &ast.FragmentSpread{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	fs.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	fs.Directives = directives

	return fs, nil
}

/*
TypeSystemExtension: https://spec.graphql.org/draft/#TypeSystemExtension

	SchemaExtension
	TypeExtension

TypeExtension: https://spec.graphql.org/draft/#TypeExtension

	ScalarTypeExtension
	ObjectTypeExtension
	InterfaceTypeExtension
	UnionTypeExtension
	EnumTypeExtension
	InputObjectTypeExtension
*/
func (p *Parser) parseTypeSystemExtension() (ast.Definition, error) {
	switch p.peekToken.Literal {
	case "schema":
		return p.parseSchemaExtension()
	case "scalar":
		return p.parseScalarTypeExtension()
	case "type":
		return p.parseObjectTypeExtension()
	case "interface":
		return p.parseInterfaceTypeExtension()
	case "union":
		return p.parseUnionTypeExtension()
	case "enum":
		return p.parseEnumTypeExtension()
	case "input":
		return p.parseInputObjectTypeExtension()
	default:
		return nil, fmt.Errorf("unexpected extension: %s", p.peekToken.Literal)
	}
}

func (p *Parser) parseInputObjectTypeExtension() (ast.Definition, error) {
	ext := &ast.InputObjectTypeExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("input"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ext.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ext.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseInputFieldsDefinition()
		if err != nil {
			return nil, err
		}
		ext.Fields = fields
	}

	return ext, nil
}

func (p *Parser) parseInputFieldsDefinition() (ast.InputFieldsDefinition, error) {
	var fields ast.InputFieldsDefinition

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		f, err := p.parseInputValueDefinition()
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return fields, nil
}

func (p *Parser) parseEnumTypeExtension() (ast.Definition, error) {
	ext := &ast.EnumTypeExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("enum"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ext.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ext.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		vals, err := p.parseEnumValuesDefinition()
		if err != nil {
			return nil, err
		}
		ext.Values = vals
	}

	return ext, nil
}

func (p *Parser) parseEnumValuesDefinition() (ast.EnumValuesDefinition, error) {
	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	var vals ast.EnumValuesDefinition
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		ev, err := p.parseEnumValueDefinition()
		if err != nil {
			return nil, err
		}
		vals = append(vals, ev)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return vals, nil
}

func (p *Parser) parseEnumValueDefinition() (*ast.EnumValueDefinition, error) {
	ev := &ast.EnumValueDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		ev.Description = desc
	}

	enumVal, err := p.parseEnumValueName()
	if err != nil {
		return nil, err
	}
	ev.Name = enumVal

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ev.Directives = directives

	return ev, nil
}

func (p *Parser) parseUnionTypeExtension() (ast.Definition, error) {
	ext := &ast.UnionTypeExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("union"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ext.Name = name

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ext.Directives = dirs

	if p.curToken.Type == token.EQUALS {
		types, err := p.parseUnionMemberTypes()
		if err != nil {
			return nil, err
		}
		ext.Types = types
	}

	if ext.Directives == nil && ext.Types == nil { // TODO: check length ?
		return nil, fmt.Errorf("unexpected: %s", p.curToken.Literal) //TODO: fix msg see https://spec.graphql.org/draft/#UnionTypeDefinition
	}

	return ext, nil
}

func (p *Parser) parseEnumValueName() (*ast.Name, error) {
	if err := p.expect(token.NAME); err != nil {
		return nil, err
	}
	if p.curToken.Literal == "true" || p.curToken.Literal == "false" || p.curToken.Literal == "null" {
		return nil, fmt.Errorf("unexpected: %s", p.curToken.Literal) // TODO: fix msg
	}
	return p.parseName()
}

func (p *Parser) parseUnionMemberTypes() (ast.UnionMemberTypes, error) {
	if err := p.expectAndAdvance(token.EQUALS); err != nil {
		return nil, err
	}
	var types ast.UnionMemberTypes
	for {
		nt, err := p.parseNamedType()
		if err != nil {
			return nil, err
		}
		types = append(types, nt)
		if p.curToken.Type == token.PIPE {
			if err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}
	return types, nil
}

func (p *Parser) parseInterfaceTypeExtension() (ast.Definition, error) {
	ext := &ast.InterfaceTypeExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("interface"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ext.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		ext.Interfaces = ii
	}

	dirs, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ext.Directives = dirs

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		ext.Fields = fields
	}

	if ext.Interfaces == nil && ext.Directives == nil && ext.Fields == nil {
		return nil, fmt.Errorf("unexpected: %s", p.curToken.Literal) //TODO: fix msg see https://spec.graphql.org/draft/#InterfaceTypeExtension
	}

	return ext, nil
}

func (p *Parser) parseSchemaExtension() (*ast.SchemaExtension, error) {
	se := &ast.SchemaExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("schema"); err != nil {
		return nil, err
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	se.Directives = directives

	if p.curToken.Type != token.LBRACE {
		if err := p.next(); err != nil {
			return nil, err
		}

		var roots []*ast.RootOperationTypeDefinition
		for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
			root, err := p.parseRootOperationTypeDefinition()
			if err != nil {
				return nil, err
			}
			roots = append(roots, root)
		}
		se.RootOperationDefs = roots

		if err := p.expectAndAdvance(token.RBRACE); err != nil {
			return nil, err
		}
	}

	return se, nil
}

func (p *Parser) parseRootOperationTypeDefinition() (*ast.RootOperationTypeDefinition, error) {
	root := &ast.RootOperationTypeDefinition{
		Position: p.curToken.Start,
	}

	opType, err := p.parseOperationType()
	if err != nil {
		return nil, err
	}
	root.OperationType = opType

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	nt, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	root.Type = nt

	return root, nil
}

func (p *Parser) parseScalarTypeExtension() (*ast.ScalarTypeExtension, error) {
	ext := &ast.ScalarTypeExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("scalar"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ext.Name = name

	dir, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	if len(dir) == 0 { // TODO: Find out if we need this
		return nil, fmt.Errorf("directives required")
	}
	ext.Directives = dir

	return ext, nil
}

func (p *Parser) parseObjectTypeExtension() (*ast.ObjectTypeExtension, error) {
	ext := &ast.ObjectTypeExtension{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("extend"); err != nil {
		return nil, err
	}
	if err := p.expectLiteralAndAdvance("type"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ext.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		ext.Interfaces = ii
	}

	dir, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ext.Directives = dir

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		ext.Fields = fields
	}

	return ext, nil
}

func (p *Parser) parseImplementsInterfaces() (ast.ImplementsInterfaces, error) {
	if err := p.expectLiteralAndAdvance("implements"); err != nil {
		return nil, err
	}
	var interfaces ast.ImplementsInterfaces
	for {
		nt, err := p.parseNamedType()
		if err != nil {
			return nil, err
		}
		interfaces = append(interfaces, nt)
		if p.curToken.Type == token.AMP {
			if err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}
	return interfaces, nil
}

func (p *Parser) parseFieldsDefinition() (ast.FieldsDefinition, error) {
	var fields ast.FieldsDefinition

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		f, err := p.parseFieldDefinition()
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return fields, nil
}

func (p *Parser) parseFieldDefinition() (*ast.FieldDefinition, error) {
	f := &ast.FieldDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		f.Description = desc
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	f.Name = name

	var args ast.ArgumentsDefinition
	if p.curToken.Type == token.LPAREN {
		args, err = p.parseArgumentsDefinition()
		if err != nil {
			return nil, err
		}
	}
	f.Arguments = args

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	f.Type = typ

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	f.Directives = directives

	return f, nil
}

func (p *Parser) parseArgumentsDefinition() (ast.ArgumentsDefinition, error) {
	var defs ast.ArgumentsDefinition

	if err := p.expectAndAdvance(token.LPAREN); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		def, err := p.parseInputValueDefinition()
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}

	if err := p.expectAndAdvance(token.RPAREN); err != nil {
		return nil, err
	}

	return defs, nil
}

func (p *Parser) parseInputValueDefinition() (*ast.InputValueDefinition, error) {
	ivd := &ast.InputValueDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		ivd.Description = desc
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	ivd.Name = name

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	ivd.Type = typ

	if p.curToken.Type == token.EQUALS {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		ivd.DefaultValue = val
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	ivd.Directives = directives

	return ivd, nil
}

func (p *Parser) parseDescription() (*ast.StringValue, error) {
	val, err := p.parseStringValue()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (p *Parser) parseStringValue() (*ast.StringValue, error) {
	if err := p.expectOneOf(token.STRING, token.BLOCK_STRING); err != nil {
		return nil, err
	}

	sv := &ast.StringValue{
		Position: p.curToken.Start,
		Value:    p.curToken.Literal,
		Block:    p.curToken.Type == token.BLOCK_STRING,
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	return sv, nil
}
