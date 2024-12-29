package ast

import "github.com/gqlhub/gqlhub-core/token"

// Node is base interface for all AST nodes.
type Node interface {
	Pos() token.Pos // Returns starting position of node.
	End() token.Pos // Returns ending position of node.
}

// Document is root node.
type Document struct {
	Position    token.Pos
	Definitions []Definition
}

func (d *Document) Pos() token.Pos { return d.Position }

// Definition can be Executable, TypeSystem, or Extension.
type Definition interface {
	Node
	definitionNode()
}

// OperationType represents type of operation.
//
// https://spec.graphql.org/draft/#OperationType
type OperationType string

const (
	OperationTypeQuery        OperationType = "query"
	OperationTypeMutation     OperationType = "mutation"
	OperationTypeSubscription OperationType = "subscription"
)

// OperationDefinition
//
// https://spec.graphql.org/draft/#OperationDefinition
type OperationDefinition struct {
	Position      token.Pos
	OperationType OperationType
	Name          *Name
	VariableDefs  VariablesDefinition
	Directives    Directives
	SelectionSet  SelectionSet
}

func (o *OperationDefinition) Pos() token.Pos  { return o.Position }
func (o *OperationDefinition) definitionNode() {}

// FragmentDefinition
//
// https://spec.graphql.org/draft/#FragmentDefinition
type FragmentDefinition struct {
	Position      token.Pos
	Name          Name
	TypeCondition NamedType
	Directives    Directives
	SelectionSet  SelectionSet
}

func (f *FragmentDefinition) Pos() token.Pos  { return f.Position }
func (f *FragmentDefinition) definitionNode() {}

// TypeDefinition covers schema, scalar, object, interface, union, enum, input.
//
// https://spec.graphql.org/draft/#TypeDefinition
type TypeDefinition interface {
	Definition
	typeSystemDefinitionNode()
}

// SchemaDefinition
//
// https://spec.graphql.org/draft/#SchemaDefinition
type SchemaDefinition struct {
	Position          token.Pos
	Description       *StringValue
	Directives        []Directive
	RootOperationDefs []RootOperationTypeDefinition
}

func (s *SchemaDefinition) Pos() token.Pos            { return s.Position }
func (s *SchemaDefinition) definitionNode()           {}
func (s *SchemaDefinition) typeSystemDefinitionNode() {}

// RootOperationTypeDefinition
//
// https://spec.graphql.org/draft/#RootOperationTypeDefinition
type RootOperationTypeDefinition struct {
	Position      token.Pos
	OperationType OperationType
	Type          NamedType
}

func (r *RootOperationTypeDefinition) Pos() token.Pos { return r.Position }

// ScalarTypeExtension
//
// https://spec.graphql.org/draft/#ScalarTypeExtension
type ScalarTypeExtension struct {
	Position   token.Pos
	Name       Name
	Directives []Directive
}

func (s *ScalarTypeExtension) Pos() token.Pos           { return s.Position }
func (s *ScalarTypeExtension) definitionNode()          {}
func (s *ScalarTypeExtension) typeSystemExtensionNode() {}

// ObjectTypeExtension
//
// https://spec.graphql.org/draft/#ObjectTypeExtension
type ObjectTypeExtension struct {
	Position   token.Pos
	Name       Name
	Interfaces []NamedType
	Directives []Directive
	Fields     []FieldDefinition
}

func (o *ObjectTypeExtension) Pos() token.Pos           { return o.Position }
func (o *ObjectTypeExtension) definitionNode()          {}
func (o *ObjectTypeExtension) typeSystemExtensionNode() {}

// TODO: add additional interfaces: InterfaceTypeExtension, UnionTypeExtension...

// FieldDefinition
//
// https://spec.graphql.org/draft/#FieldDefinition
type FieldDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Arguments   []InputValueDefinition
	Type        Type
	Directives  []Directive
}

func (f *FieldDefinition) Pos() token.Pos { return f.Position }

// InterfaceTypeDefinition
//
// https://spec.graphql.org/draft/#InterfaceTypeDefinition
type InterfaceTypeDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Interfaces  []NamedType
	Directives  []Directive
	Fields      []FieldDefinition
}

func (i *InterfaceTypeDefinition) Pos() token.Pos            { return i.Position }
func (i *InterfaceTypeDefinition) definitionNode()           {}
func (i *InterfaceTypeDefinition) typeSystemDefinitionNode() {}

// UnionTypeDefinition
//
// https://spec.graphql.org/draft/#UnionTypeDefinition
type UnionTypeDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Directives  []Directive
	Types       []NamedType
}

func (u *UnionTypeDefinition) Pos() token.Pos            { return u.Position }
func (u *UnionTypeDefinition) definitionNode()           {}
func (u *UnionTypeDefinition) typeSystemDefinitionNode() {}

// EnumTypeDefinition
//
// https://spec.graphql.org/draft/#EnumTypeDefinition
type EnumTypeDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Directives  []Directive
	Values      []EnumValueDefinition
}

func (e *EnumTypeDefinition) Pos() token.Pos            { return e.Position }
func (e *EnumTypeDefinition) definitionNode()           {}
func (e *EnumTypeDefinition) typeSystemDefinitionNode() {}

// EnumValueDefinition
//
// https://spec.graphql.org/draft/#EnumValueDefinition
type EnumValueDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Directives  []Directive
}

func (e *EnumValueDefinition) Pos() token.Pos { return e.Position }

// InputObjectTypeDefinition
//
// https://spec.graphql.org/draft/#InputObjectTypeDefinition
type InputObjectTypeDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Directives  []Directive
	Fields      []InputValueDefinition
}

func (i *InputObjectTypeDefinition) Pos() token.Pos            { return i.Position }
func (i *InputObjectTypeDefinition) definitionNode()           {}
func (i *InputObjectTypeDefinition) typeSystemDefinitionNode() {}

// InputValueDefinition
//
// https://spec.graphql.org/draft/#InputValueDefinition
type InputValueDefinition struct {
	Position     token.Pos
	Description  *StringValue
	Name         Name
	Type         Type
	DefaultValue Value
	Directives   []Directive
}

func (i *InputValueDefinition) Pos() token.Pos { return i.Position }

// DirectiveDefinition
//
// https://spec.graphql.org/draft/#DirectiveDefinition
type DirectiveDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Arguments   []InputValueDefinition
	Locations   []Name
}

func (d *DirectiveDefinition) Pos() token.Pos            { return d.Position }
func (d *DirectiveDefinition) definitionNode()           {}
func (d *DirectiveDefinition) typeSystemDefinitionNode() {}

// TypeSystemExtension
//
// https://spec.graphql.org/draft/#TypeSystemExtension
type TypeSystemExtension interface {
	Definition
	typeSystemExtensionNode()
}

// SchemaExtension
//
// https://spec.graphql.org/draft/#SchemaExtension
type SchemaExtension struct {
	Position          token.Pos
	Directives        []Directive
	RootOperationDefs []RootOperationTypeDefinition
}

func (s *SchemaExtension) Pos() token.Pos           { return s.Position }
func (s *SchemaExtension) definitionNode()          {}
func (s *SchemaExtension) typeSystemExtensionNode() {}

// ScalarTypeDefinition
//
// https://spec.graphql.org/draft/#ScalarTypeDefinition
type ScalarTypeDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Directives  []Directive
}

func (s *ScalarTypeDefinition) Pos() token.Pos            { return s.Position }
func (s *ScalarTypeDefinition) definitionNode()           {}
func (s *ScalarTypeDefinition) typeSystemDefinitionNode() {}

// ObjectTypeDefinition
//
// https://spec.graphql.org/draft/#ObjectTypeDefinition
type ObjectTypeDefinition struct {
	Position    token.Pos
	Description *StringValue
	Name        Name
	Interfaces  []NamedType
	Directives  []Directive
	Fields      []FieldDefinition
}

func (o *ObjectTypeDefinition) Pos() token.Pos            { return o.Position }
func (o *ObjectTypeDefinition) definitionNode()           {}
func (o *ObjectTypeDefinition) typeSystemDefinitionNode() {}

// SelectionSet
//
// https://spec.graphql.org/draft/#SelectionSet
type SelectionSet struct {
	Position   token.Pos
	Selections []Selection
}

func (s *SelectionSet) Pos() token.Pos { return s.Position }

// Selection can be Field, FragmentSpread, InlineFragment
//
// https://spec.graphql.org/draft/#Selection
type Selection interface {
	Node
	selectionNode()
}

// Field
//
// https://spec.graphql.org/draft/#Field
type Field struct {
	Position     token.Pos
	Alias        *Name
	Name         Name
	Arguments    Arguments
	Directives   Directives
	SelectionSet *SelectionSet
}

func (f *Field) Pos() token.Pos { return f.Position }
func (f *Field) selectionNode() {}

// FragmentSpread
//
// https://spec.graphql.org/draft/#FragmentSpread
type FragmentSpread struct {
	Position   token.Pos
	Name       Name
	Directives Directives
}

func (fs *FragmentSpread) Pos() token.Pos { return fs.Position }
func (fs *FragmentSpread) selectionNode() {}

// InlineFragment
//
// https://spec.graphql.org/draft/#InlineFragment
type InlineFragment struct {
	Position      token.Pos
	TypeCondition *NamedType
	Directives    Directives
	SelectionSet  SelectionSet
}

func (inf *InlineFragment) Pos() token.Pos { return inf.Position }
func (inf *InlineFragment) selectionNode() {}

// Directive
//
// https://spec.graphql.org/draft/#Directive
type Directive struct {
	Position  token.Pos
	Name      Name
	Arguments Arguments
}

// Directives
//
// https://spec.graphql.org/draft/#Directives
type Directives = []Directive

func (d *Directive) Pos() token.Pos { return d.Position }

// Arguments
//
// https://spec.graphql.org/draft/#Arguments
type Arguments = []Argument

// Argument
//
// https://spec.graphql.org/draft/#Argument
type Argument struct {
	Position token.Pos
	Name     Name
	Value    Value
}

func (a *Argument) Pos() token.Pos { return a.Position }

// Value can be IntValue, FloatValue, StringValue, BooleanValue,
// NullValue, EnumValue, ListValue, ObjectValue, Variable.
//
// https://spec.graphql.org/draft/#Value
type Value interface {
	Node
	valueNode()
}

// IntValue
//
// https://spec.graphql.org/draft/#IntValue
type IntValue struct {
	Position token.Pos
	Value    string
}

func (v *IntValue) Pos() token.Pos { return v.Position }
func (v *IntValue) valueNode()     {}

// FloatValue
//
// https://spec.graphql.org/draft/#FloatValue
type FloatValue struct {
	Position token.Pos
	Value    string
}

func (v *FloatValue) Pos() token.Pos { return v.Position }
func (v *FloatValue) valueNode()     {}

// StringValue
//
// https://spec.graphql.org/draft/#StringValue
type StringValue struct {
	Position token.Pos
	Value    string
	Block    bool
}

func (v *StringValue) Pos() token.Pos { return v.Position }
func (v *StringValue) valueNode()     {}

// BooleanValue
//
// https://spec.graphql.org/draft/#BooleanValue
type BooleanValue struct {
	Position token.Pos
	Value    bool
}

func (v *BooleanValue) Pos() token.Pos { return v.Position }
func (v *BooleanValue) valueNode()     {}

// NullValue
//
// https://spec.graphql.org/draft/#NullValue
type NullValue struct {
	Position token.Pos
}

func (v *NullValue) Pos() token.Pos { return v.Position }
func (v *NullValue) valueNode()     {}

// EnumValue
//
// https://spec.graphql.org/draft/#EnumValue
type EnumValue struct {
	Position token.Pos
	Value    string
}

func (v *EnumValue) Pos() token.Pos { return v.Position }
func (v *EnumValue) valueNode()     {}

// ListValue
//
// https://spec.graphql.org/draft/#ListValue
type ListValue struct {
	Position token.Pos
	Values   []Value
}

func (v *ListValue) Pos() token.Pos { return v.Position }
func (v *ListValue) valueNode()     {}

// ObjectValue
//
// https://spec.graphql.org/draft/#ObjectValue
type ObjectValue struct {
	Position token.Pos
	Fields   []ObjectField
}

func (v *ObjectValue) Pos() token.Pos { return v.Position }
func (v *ObjectValue) valueNode()     {}

// ObjectField
//
// https://spec.graphql.org/draft/#ObjectField
type ObjectField struct {
	Position token.Pos
	Name     Name
	Value    Value
}

func (o *ObjectField) Pos() token.Pos { return o.Position }

// Variable
//
// https://spec.graphql.org/draft/#Variable
type Variable struct {
	Position token.Pos
	Name     Name
}

func (v *Variable) Pos() token.Pos { return v.Position }
func (v *Variable) valueNode()     {}

// VariablesDefinition
//
// https://spec.graphql.org/draft/#VariablesDefinition
type VariablesDefinition = []VariableDefinition

// VariableDefinition
//
// https://spec.graphql.org/draft/#VariableDefinition
type VariableDefinition struct {
	Position     token.Pos
	Variable     Variable
	Type         Type
	DefaultValue Value
	Directives   Directives
}

func (vd *VariableDefinition) Pos() token.Pos { return vd.Position }

// Type can be NamedType, ListType, NonNullType.
//
// https://spec.graphql.org/draft/#Type
type Type interface {
	Node
	typeNode()
}

// NamedType
//
// https://spec.graphql.org/draft/#NamedType
type NamedType struct {
	Position token.Pos
	Name     Name
}

func (n *NamedType) Pos() token.Pos { return n.Position }
func (n *NamedType) typeNode()      {}

// ListType
//
// https://spec.graphql.org/draft/#ListType
type ListType struct {
	Position token.Pos
	Type     Type
}

func (l *ListType) Pos() token.Pos { return l.Position }
func (l *ListType) typeNode()      {}

// NonNullType
//
// https://spec.graphql.org/draft/#NonNullType
type NonNullType struct {
	Position token.Pos
	Type     Type
}

func (n *NonNullType) Pos() token.Pos { return n.Position }
func (n *NonNullType) typeNode()      {}

// Name
//
// https://spec.graphql.org/draft/#Name
type Name struct {
	Position token.Pos
	Value    string
}

func (n *Name) Pos() token.Pos { return n.Position }
