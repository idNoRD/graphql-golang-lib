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
		doc.Definitions = append(doc.Definitions, def)
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
		return p.parseAnonymousOperationDefinition()
	}

	var tok token.Token
	if isDescription(p.curToken.Type) {
		tok = p.peekToken
	} else {
		tok = p.curToken
	}

	if tok.Type == token.NAME {
		switch tok.Literal {
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
	schemaDef := &ast.SchemaDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		schemaDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("schema"); err != nil {
		return nil, err
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	schemaDef.Directives = directives

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		def, err := p.parseRootOperationTypeDefinition()
		if err != nil {
			return nil, err
		}
		schemaDef.RootOperationDefs = append(schemaDef.RootOperationDefs, def)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return schemaDef, nil
}

func (p *Parser) parseScalarTypeDefinition() (ast.Definition, error) {
	scalarTypeDef := &ast.ScalarTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		scalarTypeDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("scalar"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	scalarTypeDef.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	scalarTypeDef.Directives = directives

	return scalarTypeDef, nil
}

func (p *Parser) parseObjectTypeDefinition() (ast.Definition, error) {
	objTypeDef := &ast.ObjectTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		objTypeDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("type"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	objTypeDef.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		objTypeDef.Interfaces = ii
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	objTypeDef.Directives = directives

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		objTypeDef.Fields = fields
	}

	return objTypeDef, nil
}

func (p *Parser) parseInterfaceTypeDefinition() (ast.Definition, error) {
	interfaceTypeDef := &ast.InterfaceTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		interfaceTypeDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("interface"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	interfaceTypeDef.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		interfaceTypeDef.Interfaces = ii
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	interfaceTypeDef.Directives = directives

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		interfaceTypeDef.Fields = fields
	}

	return interfaceTypeDef, nil
}

func (p *Parser) parseUnionTypeDefinition() (ast.Definition, error) {
	unionTypeDef := &ast.UnionTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		unionTypeDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("union"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	unionTypeDef.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	unionTypeDef.Directives = directives

	if p.curToken.Type == token.EQUALS {
		types, err := p.parseUnionMemberTypes()
		if err != nil {
			return nil, err
		}
		unionTypeDef.Types = types
	}

	return unionTypeDef, nil
}

func (p *Parser) parseEnumTypeDefinition() (ast.Definition, error) {
	enumTypeDef := &ast.EnumTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		enumTypeDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("enum"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	enumTypeDef.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	enumTypeDef.Directives = directives

	if p.curToken.Type == token.LBRACE {
		vals, err := p.parseEnumValuesDefinition()
		if err != nil {
			return nil, err
		}
		enumTypeDef.Values = vals
	}

	return enumTypeDef, nil
}

func (p *Parser) parseInputObjectTypeDefinition() (ast.Definition, error) {
	inputObjTypeDef := &ast.InputObjectTypeDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		inputObjTypeDef.Description = desc
	}

	if err := p.expectLiteralAndAdvance("input"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	inputObjTypeDef.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	inputObjTypeDef.Directives = directives

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseInputFieldsDefinition()
		if err != nil {
			return nil, err
		}
		inputObjTypeDef.Fields = fields
	}

	return inputObjTypeDef, nil
}

func (p *Parser) parseDirectiveDefinition() (ast.Definition, error) {
	directiveDef := &ast.DirectiveDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		directiveDef.Description = desc
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
	directiveDef.Name = name

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArgumentsDefinition()
		if err != nil {
			return nil, err
		}
		directiveDef.Arguments = args
	}

	if p.curToken.Literal == "repeatable" {
		if err := p.next(); err != nil {
			return nil, err
		}
		directiveDef.Repeatable = true
	}

	if err := p.expectLiteralAndAdvance("on"); err != nil {
		return nil, err
	}

	locations, err := p.parseDirectiveLocations()
	if err != nil {
		return nil, err
	}
	directiveDef.Locations = locations

	return directiveDef, nil
}

func (p *Parser) parseDirectiveLocations() ([]*ast.Name, error) {
	var directiveLocations []*ast.Name
	for {
		nt, err := p.parseName()
		if err != nil {
			return nil, err
		}
		directiveLocations = append(directiveLocations, nt)
		if p.curToken.Type == token.PIPE {
			if err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}
	return directiveLocations, nil
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
	opDef := &ast.OperationDefinition{
		Position: p.curToken.Start,
	}

	opType, err := p.parseOperationType()
	if err != nil {
		return nil, err
	}
	opDef.OperationType = opType

	if p.curToken.Type == token.NAME {
		name, err := p.parseName()
		if err != nil {
			return nil, err
		}
		opDef.Name = name
	}

	if p.curToken.Type == token.LPAREN {
		varDefs, err := p.parseVariableDefinitions()
		if err != nil {
			return nil, err
		}
		opDef.VariableDefs = varDefs
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	opDef.Directives = directives

	selectionSet, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	opDef.SelectionSet = selectionSet

	return opDef, nil
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
	var varDefs []*ast.VariableDefinition

	if err := p.expectAndAdvance(token.LPAREN); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		def, err := p.parseVariableDefinition()
		if err != nil {
			return nil, err
		}
		varDefs = append(varDefs, def)
	}

	if err := p.expectAndAdvance(token.RPAREN); err != nil {
		return nil, err
	}

	return varDefs, nil
}

func (p *Parser) parseVariableDefinition() (*ast.VariableDefinition, error) {
	varDef := &ast.VariableDefinition{
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
	varDef.Variable = variable

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	varDef.Type = typ

	if p.curToken.Type == token.EQUALS {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		varDef.DefaultValue = val
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	varDef.Directives = directives

	return varDef, nil
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

func (p *Parser) parseAnonymousOperationDefinition() (*ast.OperationDefinition, error) {
	opDef := &ast.OperationDefinition{
		OperationType: ast.OperationTypeQuery,
	}

	selectionSet, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	opDef.SelectionSet = selectionSet

	return opDef, nil
}

func (p *Parser) parseFragmentDefinition() (*ast.FragmentDefinition, error) {
	fragmentDef := &ast.FragmentDefinition{
		Position: p.curToken.Start,
	}

	if err := p.expectLiteralAndAdvance("fragment"); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	fragmentDef.Name = name

	typeCond, err := p.parseTypeCondition()
	if err != nil {
		return nil, err
	}
	fragmentDef.TypeCondition = typeCond

	selectionSet, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	fragmentDef.SelectionSet = selectionSet

	return fragmentDef, nil
}

func (p *Parser) parseTypeCondition() (*ast.NamedType, error) {
	if err := p.expectLiteralAndAdvance("on"); err != nil {
		return nil, err
	}
	return p.parseNamedType()
}

func (p *Parser) parseSelectionSet() (*ast.SelectionSet, error) {
	selectionSet := &ast.SelectionSet{
		Position: p.curToken.Start,
	}

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		selection, err := p.parseSelection()
		if err != nil {
			return nil, err
		}
		selectionSet.Selections = append(selectionSet.Selections, selection)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return selectionSet, nil
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
	inlineFragment := &ast.InlineFragment{
		Position: p.curToken.Start,
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	nt, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	inlineFragment.TypeCondition = nt

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	inlineFragment.Directives = directives

	selectionSet, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	inlineFragment.SelectionSet = selectionSet

	return inlineFragment, nil
}

func (p *Parser) parseNamedType() (*ast.NamedType, error) {
	namedType := &ast.NamedType{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	namedType.Name = name

	return namedType, nil
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
	directive := &ast.Directive{
		Position: p.curToken.Start,
	}

	if err := p.expectAndAdvance(token.AT); err != nil {
		return nil, err
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	directive.Name = name

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArguments()
		if err != nil {
			return nil, err
		}
		directive.Arguments = args
	}

	return directive, nil
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
	listValue := &ast.ListValue{
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
		listValue.Values = append(listValue.Values, val)
	}

	if err := p.expectAndAdvance(token.RBRACK); err != nil {
		return nil, err
	}

	return listValue, nil
}

func (p *Parser) parseObjectValue() (ast.Value, error) {
	objValue := &ast.ObjectValue{
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
		objValue.Fields = append(objValue.Fields, field)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return objValue, nil
}

func (p *Parser) parseObjectField() (*ast.ObjectField, error) {
	objField := &ast.ObjectField{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	objField.Name = name

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	objField.Value = val

	return objField, nil
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
	fragmentSpread := &ast.FragmentSpread{
		Position: p.curToken.Start,
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	fragmentSpread.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	fragmentSpread.Directives = directives

	return fragmentSpread, nil
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
	inputObjTypeExtension := &ast.InputObjectTypeExtension{
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
	inputObjTypeExtension.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	inputObjTypeExtension.Directives = directives

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseInputFieldsDefinition()
		if err != nil {
			return nil, err
		}
		inputObjTypeExtension.Fields = fields
	}

	return inputObjTypeExtension, nil
}

func (p *Parser) parseInputFieldsDefinition() ([]*ast.InputValueDefinition, error) {
	var inputValueDef []*ast.InputValueDefinition

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		f, err := p.parseInputValueDefinition()
		if err != nil {
			return nil, err
		}
		inputValueDef = append(inputValueDef, f)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return inputValueDef, nil
}

func (p *Parser) parseEnumTypeExtension() (ast.Definition, error) {
	enumTypeExtension := &ast.EnumTypeExtension{
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
	enumTypeExtension.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	enumTypeExtension.Directives = directives

	if p.curToken.Type == token.LBRACE {
		vals, err := p.parseEnumValuesDefinition()
		if err != nil {
			return nil, err
		}
		enumTypeExtension.Values = vals
	}

	return enumTypeExtension, nil
}

func (p *Parser) parseEnumValuesDefinition() ([]*ast.EnumValueDefinition, error) {
	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	var vals []*ast.EnumValueDefinition
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
	unionTypeExtension := &ast.UnionTypeExtension{
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
	unionTypeExtension.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	unionTypeExtension.Directives = directives

	if p.curToken.Type == token.EQUALS {
		types, err := p.parseUnionMemberTypes()
		if err != nil {
			return nil, err
		}
		unionTypeExtension.Types = types
	}

	if unionTypeExtension.Directives == nil && unionTypeExtension.Types == nil { // TODO: check length ?
		return nil, fmt.Errorf("unexpected: %s", p.curToken.Literal) //TODO: fix msg see https://spec.graphql.org/draft/#UnionTypeDefinition
	}

	return unionTypeExtension, nil
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

func (p *Parser) parseUnionMemberTypes() ([]*ast.NamedType, error) {
	if err := p.expectAndAdvance(token.EQUALS); err != nil {
		return nil, err
	}
	var types []*ast.NamedType
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
	interfaceTypeExtension := &ast.InterfaceTypeExtension{
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
	interfaceTypeExtension.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		interfaceTypeExtension.Interfaces = ii
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	interfaceTypeExtension.Directives = directives

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		interfaceTypeExtension.Fields = fields
	}

	if interfaceTypeExtension.Interfaces == nil && interfaceTypeExtension.Directives == nil && interfaceTypeExtension.Fields == nil {
		return nil, fmt.Errorf("unexpected: %s", p.curToken.Literal) //TODO: fix msg see https://spec.graphql.org/draft/#InterfaceTypeExtension
	}

	return interfaceTypeExtension, nil
}

func (p *Parser) parseSchemaExtension() (*ast.SchemaExtension, error) {
	schemaExtension := &ast.SchemaExtension{
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
	schemaExtension.Directives = directives

	if p.curToken.Type != token.LBRACE {
		if err := p.next(); err != nil {
			return nil, err
		}

		for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
			def, err := p.parseRootOperationTypeDefinition()
			if err != nil {
				return nil, err
			}
			schemaExtension.RootOperationDefs = append(schemaExtension.RootOperationDefs, def)
		}

		if err := p.expectAndAdvance(token.RBRACE); err != nil {
			return nil, err
		}
	}

	return schemaExtension, nil
}

func (p *Parser) parseRootOperationTypeDefinition() (*ast.RootOperationTypeDefinition, error) {
	rootOpTypeDef := &ast.RootOperationTypeDefinition{
		Position: p.curToken.Start,
	}

	opType, err := p.parseOperationType()
	if err != nil {
		return nil, err
	}
	rootOpTypeDef.OperationType = opType

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	namedType, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	rootOpTypeDef.Type = namedType

	return rootOpTypeDef, nil
}

func (p *Parser) parseScalarTypeExtension() (*ast.ScalarTypeExtension, error) {
	scalarTypeExtension := &ast.ScalarTypeExtension{
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
	scalarTypeExtension.Name = name

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	if len(directives) == 0 { // TODO: Find out if we need this
		return nil, fmt.Errorf("directives required")
	}
	scalarTypeExtension.Directives = directives

	return scalarTypeExtension, nil
}

func (p *Parser) parseObjectTypeExtension() (*ast.ObjectTypeExtension, error) {
	objTypeExtension := &ast.ObjectTypeExtension{
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
	objTypeExtension.Name = name

	if p.curToken.Literal == "implements" {
		ii, err := p.parseImplementsInterfaces()
		if err != nil {
			return nil, err
		}
		objTypeExtension.Interfaces = ii
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	objTypeExtension.Directives = directives

	if p.curToken.Type == token.LBRACE {
		fields, err := p.parseFieldsDefinition()
		if err != nil {
			return nil, err
		}
		objTypeExtension.Fields = fields
	}

	return objTypeExtension, nil
}

func (p *Parser) parseImplementsInterfaces() ([]*ast.NamedType, error) {
	if err := p.expectLiteralAndAdvance("implements"); err != nil {
		return nil, err
	}
	var interfaces []*ast.NamedType
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

func (p *Parser) parseFieldsDefinition() ([]*ast.FieldDefinition, error) {
	var fieldsDef []*ast.FieldDefinition

	if err := p.expectAndAdvance(token.LBRACE); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		f, err := p.parseFieldDefinition()
		if err != nil {
			return nil, err
		}
		fieldsDef = append(fieldsDef, f)
	}

	if err := p.expectAndAdvance(token.RBRACE); err != nil {
		return nil, err
	}

	return fieldsDef, nil
}

func (p *Parser) parseFieldDefinition() (*ast.FieldDefinition, error) {
	fieldDef := &ast.FieldDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		fieldDef.Description = desc
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	fieldDef.Name = name

	if p.curToken.Type == token.LPAREN {
		args, err := p.parseArgumentsDefinition()
		if err != nil {
			return nil, err
		}
		fieldDef.Arguments = args
	}

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	fieldDef.Type = typ

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	fieldDef.Directives = directives

	return fieldDef, nil
}

func (p *Parser) parseArgumentsDefinition() ([]*ast.InputValueDefinition, error) {
	var argumentsDef []*ast.InputValueDefinition

	if err := p.expectAndAdvance(token.LPAREN); err != nil {
		return nil, err
	}

	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		def, err := p.parseInputValueDefinition()
		if err != nil {
			return nil, err
		}
		argumentsDef = append(argumentsDef, def)
	}

	if err := p.expectAndAdvance(token.RPAREN); err != nil {
		return nil, err
	}

	return argumentsDef, nil
}

func (p *Parser) parseInputValueDefinition() (*ast.InputValueDefinition, error) {
	inputValueDef := &ast.InputValueDefinition{
		Position: p.curToken.Start,
	}

	if isDescription(p.curToken.Type) {
		desc, err := p.parseDescription()
		if err != nil {
			return nil, err
		}
		inputValueDef.Description = desc
	}

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	inputValueDef.Name = name

	if err := p.expectAndAdvance(token.COLON); err != nil {
		return nil, err
	}

	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	inputValueDef.Type = typ

	if p.curToken.Type == token.EQUALS {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		inputValueDef.DefaultValue = val
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	inputValueDef.Directives = directives

	return inputValueDef, nil
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

	strValue := &ast.StringValue{
		Position: p.curToken.Start,
		Value:    p.curToken.Literal,
		Block:    p.curToken.Type == token.BLOCK_STRING,
	}

	if err := p.next(); err != nil {
		return nil, err
	}

	return strValue, nil
}
