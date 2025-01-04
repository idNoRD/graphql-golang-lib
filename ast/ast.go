package ast

// Node is base interface for all AST nodes.
type Node interface {
	Pos() int // Returns starting position of node.
	End() int // Returns ending position of node.
}

// Document is root node.
type Document struct {
	Definitions []Definition
}

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
	Position      int
	OperationType OperationType
	Name          *Name
	VariableDefs  []*VariableDefinition
	Directives    []*Directive
	SelectionSet  *SelectionSet
}

func (o *OperationDefinition) Pos() int        { return o.Position }
func (o *OperationDefinition) End() int        { return -1 }
func (o *OperationDefinition) definitionNode() {}

// FragmentDefinition
//
// https://spec.graphql.org/draft/#FragmentDefinition
type FragmentDefinition struct {
	Position      int
	Name          *Name
	TypeCondition *NamedType
	Directives    []*Directive
	SelectionSet  *SelectionSet
}

func (f *FragmentDefinition) Pos() int        { return f.Position }
func (f *FragmentDefinition) End() int        { return -1 }
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
	Position          int
	Description       *StringValue
	Directives        []*Directive
	RootOperationDefs []*RootOperationTypeDefinition
}

func (s *SchemaDefinition) Pos() int                  { return s.Position }
func (s *SchemaDefinition) End() int                  { return -1 }
func (s *SchemaDefinition) definitionNode()           {}
func (s *SchemaDefinition) typeSystemDefinitionNode() {}

// RootOperationTypeDefinition
//
// https://spec.graphql.org/draft/#RootOperationTypeDefinition
type RootOperationTypeDefinition struct {
	Position      int
	OperationType OperationType
	Type          *NamedType
}

func (r *RootOperationTypeDefinition) Pos() int { return r.Position }
func (r *RootOperationTypeDefinition) End() int { return -1 }

// ScalarTypeExtension
//
// https://spec.graphql.org/draft/#ScalarTypeExtension
type ScalarTypeExtension struct {
	Position   int
	Name       *Name
	Directives []*Directive
}

func (s *ScalarTypeExtension) Pos() int                 { return s.Position }
func (s *ScalarTypeExtension) End() int                 { return -1 }
func (s *ScalarTypeExtension) definitionNode()          {}
func (s *ScalarTypeExtension) typeSystemExtensionNode() {}

// ObjectTypeExtension
//
// https://spec.graphql.org/draft/#ObjectTypeExtension
type ObjectTypeExtension struct {
	Position   int
	Name       *Name
	Interfaces []*NamedType
	Directives []*Directive
	Fields     []*FieldDefinition
}

func (o *ObjectTypeExtension) Pos() int                 { return o.Position }
func (o *ObjectTypeExtension) End() int                 { return -1 }
func (o *ObjectTypeExtension) definitionNode()          {}
func (o *ObjectTypeExtension) typeSystemExtensionNode() {}

// InterfaceTypeExtension
//
// https://spec.graphql.org/draft/#InterfaceTypeExtension
type InterfaceTypeExtension struct {
	Position   int
	Name       *Name
	Interfaces []*NamedType
	Directives []*Directive
	Fields     []*FieldDefinition
}

func (i *InterfaceTypeExtension) Pos() int                 { return i.Position }
func (i *InterfaceTypeExtension) End() int                 { return -1 }
func (i *InterfaceTypeExtension) definitionNode()          {}
func (i *InterfaceTypeExtension) typeSystemExtensionNode() {}

// UnionTypeExtension
//
// https://spec.graphql.org/draft/#UnionTypeExtension
type UnionTypeExtension struct {
	Position   int
	Name       *Name
	Directives []*Directive
	Types      []*NamedType
}

func (u *UnionTypeExtension) Pos() int                 { return u.Position }
func (u *UnionTypeExtension) End() int                 { return -1 }
func (u *UnionTypeExtension) definitionNode()          {}
func (u *UnionTypeExtension) typeSystemExtensionNode() {}

// EnumTypeExtension
//
// https://spec.graphql.org/draft/#EnumTypeExtension
type EnumTypeExtension struct {
	Position   int
	Name       *Name
	Directives []*Directive
	Values     []*EnumValueDefinition
}

func (e *EnumTypeExtension) Pos() int                 { return e.Position }
func (e *EnumTypeExtension) End() int                 { return -1 }
func (e *EnumTypeExtension) definitionNode()          {}
func (e *EnumTypeExtension) typeSystemExtensionNode() {}

// InputObjectTypeExtension
//
// https://spec.graphql.org/draft/#InputObjectTypeExtension
type InputObjectTypeExtension struct {
	Position   int
	Name       *Name
	Directives []*Directive
	Fields     []*InputValueDefinition
}

func (i *InputObjectTypeExtension) Pos() int                 { return i.Position }
func (i *InputObjectTypeExtension) End() int                 { return -1 }
func (i *InputObjectTypeExtension) definitionNode()          {}
func (i *InputObjectTypeExtension) typeSystemExtensionNode() {}

// FieldDefinition
//
// https://spec.graphql.org/draft/#FieldDefinition
type FieldDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Arguments   []*InputValueDefinition
	Type        Type
	Directives  []*Directive
}

func (f *FieldDefinition) Pos() int { return f.Position }
func (f *FieldDefinition) End() int { return -1 }

// InterfaceTypeDefinition
//
// https://spec.graphql.org/draft/#InterfaceTypeDefinition
type InterfaceTypeDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Interfaces  []*NamedType
	Directives  []*Directive
	Fields      []*FieldDefinition
}

func (i *InterfaceTypeDefinition) Pos() int                  { return i.Position }
func (i *InterfaceTypeDefinition) End() int                  { return -1 }
func (i *InterfaceTypeDefinition) definitionNode()           {}
func (i *InterfaceTypeDefinition) typeSystemDefinitionNode() {}

// UnionTypeDefinition
//
// https://spec.graphql.org/draft/#UnionTypeDefinition
type UnionTypeDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Directives  []*Directive
	Types       []*NamedType
}

func (u *UnionTypeDefinition) Pos() int                  { return u.Position }
func (u *UnionTypeDefinition) End() int                  { return -1 }
func (u *UnionTypeDefinition) definitionNode()           {}
func (u *UnionTypeDefinition) typeSystemDefinitionNode() {}

// EnumTypeDefinition
//
// https://spec.graphql.org/draft/#EnumTypeDefinition
type EnumTypeDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Directives  []*Directive
	Values      []*EnumValueDefinition
}

func (e *EnumTypeDefinition) Pos() int                  { return e.Position }
func (e *EnumTypeDefinition) End() int                  { return -1 }
func (e *EnumTypeDefinition) definitionNode()           {}
func (e *EnumTypeDefinition) typeSystemDefinitionNode() {}

// EnumValueDefinition
//
// https://spec.graphql.org/draft/#EnumValueDefinition
type EnumValueDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Directives  []*Directive
}

func (e *EnumValueDefinition) Pos() int { return e.Position }
func (e *EnumValueDefinition) End() int { return -1 }

// InputObjectTypeDefinition
//
// https://spec.graphql.org/draft/#InputObjectTypeDefinition
type InputObjectTypeDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Directives  []*Directive
	Fields      []*InputValueDefinition
}

func (i *InputObjectTypeDefinition) Pos() int                  { return i.Position }
func (i *InputObjectTypeDefinition) End() int                  { return -1 }
func (i *InputObjectTypeDefinition) definitionNode()           {}
func (i *InputObjectTypeDefinition) typeSystemDefinitionNode() {}

// InputValueDefinition
//
// https://spec.graphql.org/draft/#InputValueDefinition
type InputValueDefinition struct {
	Position     int
	Description  *StringValue
	Name         *Name
	Type         Type
	DefaultValue Value
	Directives   []*Directive
}

func (i *InputValueDefinition) Pos() int { return i.Position }
func (i *InputValueDefinition) End() int { return -1 }

// DirectiveDefinition
//
// https://spec.graphql.org/draft/#DirectiveDefinition
type DirectiveDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Arguments   []*InputValueDefinition
	Repeatable  bool
	Locations   []*Name
}

func (d *DirectiveDefinition) Pos() int                  { return d.Position }
func (d *DirectiveDefinition) End() int                  { return -1 }
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
	Position          int
	Directives        []*Directive
	RootOperationDefs []*RootOperationTypeDefinition
}

func (s *SchemaExtension) Pos() int                 { return s.Position }
func (s *SchemaExtension) End() int                 { return -1 }
func (s *SchemaExtension) definitionNode()          {}
func (s *SchemaExtension) typeSystemExtensionNode() {}

// ScalarTypeDefinition
//
// https://spec.graphql.org/draft/#ScalarTypeDefinition
type ScalarTypeDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Directives  []*Directive
}

func (s *ScalarTypeDefinition) Pos() int                  { return s.Position }
func (s *ScalarTypeDefinition) End() int                  { return -1 }
func (s *ScalarTypeDefinition) definitionNode()           {}
func (s *ScalarTypeDefinition) typeSystemDefinitionNode() {}

// ObjectTypeDefinition
//
// https://spec.graphql.org/draft/#ObjectTypeDefinition
type ObjectTypeDefinition struct {
	Position    int
	Description *StringValue
	Name        *Name
	Interfaces  []*NamedType
	Directives  []*Directive
	Fields      []*FieldDefinition
}

func (o *ObjectTypeDefinition) Pos() int                  { return o.Position }
func (o *ObjectTypeDefinition) End() int                  { return -1 }
func (o *ObjectTypeDefinition) definitionNode()           {}
func (o *ObjectTypeDefinition) typeSystemDefinitionNode() {}

// SelectionSet
//
// https://spec.graphql.org/draft/#SelectionSet
type SelectionSet struct {
	Position   int
	Selections []Selection
}

func (s *SelectionSet) Pos() int { return s.Position }
func (s *SelectionSet) End() int { return -1 }

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
	Position     int
	Alias        *Name
	Name         *Name
	Arguments    []*Argument
	Directives   []*Directive
	SelectionSet *SelectionSet
}

func (f *Field) Pos() int       { return f.Position }
func (f *Field) End() int       { return -1 }
func (f *Field) selectionNode() {}

// FragmentSpread
//
// https://spec.graphql.org/draft/#FragmentSpread
type FragmentSpread struct {
	Position   int
	Name       *Name
	Directives []*Directive
}

func (fs *FragmentSpread) Pos() int       { return fs.Position }
func (fs *FragmentSpread) End() int       { return -1 }
func (fs *FragmentSpread) selectionNode() {}

// InlineFragment
//
// https://spec.graphql.org/draft/#InlineFragment
type InlineFragment struct {
	Position      int
	TypeCondition *NamedType
	Directives    []*Directive
	SelectionSet  *SelectionSet
}

func (inf *InlineFragment) Pos() int       { return inf.Position }
func (inf *InlineFragment) End() int       { return -1 }
func (inf *InlineFragment) selectionNode() {}

// Directive
//
// https://spec.graphql.org/draft/#Directive
type Directive struct {
	Position  int
	Name      *Name
	Arguments []*Argument
}

//// Directives
////
//// https://spec.graphql.org/draft/#Directives
//type Directives = []Directive

func (d *Directive) Pos() int { return d.Position }
func (d *Directive) End() int { return -1 }

// Argument
//
// https://spec.graphql.org/draft/#Argument
type Argument struct {
	Position int
	Name     *Name
	Value    Value
}

func (a *Argument) Pos() int { return a.Position }
func (a *Argument) End() int { return -1 }

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
	Position int
	Value    string
}

func (v *IntValue) Pos() int   { return v.Position }
func (v *IntValue) End() int   { return -1 }
func (v *IntValue) valueNode() {}

// FloatValue
//
// https://spec.graphql.org/draft/#FloatValue
type FloatValue struct {
	Position int
	Value    string
}

func (v *FloatValue) Pos() int   { return v.Position }
func (v *FloatValue) End() int   { return -1 }
func (v *FloatValue) valueNode() {}

// StringValue
//
// https://spec.graphql.org/draft/#StringValue
type StringValue struct {
	Position int
	Value    string
	Block    bool
}

func (v *StringValue) Pos() int   { return v.Position }
func (v *StringValue) End() int   { return -1 }
func (v *StringValue) valueNode() {}

// BooleanValue
//
// https://spec.graphql.org/draft/#BooleanValue
type BooleanValue struct {
	Position int
	Value    bool
}

func (v *BooleanValue) Pos() int   { return v.Position }
func (v *BooleanValue) End() int   { return -1 }
func (v *BooleanValue) valueNode() {}

// NullValue
//
// https://spec.graphql.org/draft/#NullValue
type NullValue struct {
	Position int
}

func (v *NullValue) Pos() int   { return v.Position }
func (v *NullValue) End() int   { return -1 }
func (v *NullValue) valueNode() {}

// EnumValue
//
// https://spec.graphql.org/draft/#EnumValue
type EnumValue struct {
	Position int
	Value    string
}

func (v *EnumValue) Pos() int   { return v.Position }
func (v *EnumValue) End() int   { return -1 }
func (v *EnumValue) valueNode() {}

// ListValue
//
// https://spec.graphql.org/draft/#ListValue
type ListValue struct {
	Position int
	Values   []Value
}

func (v *ListValue) Pos() int   { return v.Position }
func (v *ListValue) End() int   { return -1 }
func (v *ListValue) valueNode() {}

// ObjectValue
//
// https://spec.graphql.org/draft/#ObjectValue
type ObjectValue struct {
	Position int
	Fields   []*ObjectField
}

func (v *ObjectValue) Pos() int   { return v.Position }
func (v *ObjectValue) End() int   { return -1 }
func (v *ObjectValue) valueNode() {}

// ObjectField
//
// https://spec.graphql.org/draft/#ObjectField
type ObjectField struct {
	Position int
	Name     *Name
	Value    Value
}

func (o *ObjectField) Pos() int { return o.Position }
func (o *ObjectField) End() int { return -1 }

// Variable
//
// https://spec.graphql.org/draft/#Variable
type Variable struct {
	Position int
	Name     *Name
}

func (v *Variable) Pos() int   { return v.Position }
func (v *Variable) End() int   { return -1 }
func (v *Variable) valueNode() {}

// VariableDefinition
//
// https://spec.graphql.org/draft/#VariableDefinition
type VariableDefinition struct {
	Position     int
	Variable     *Variable
	Type         Type
	DefaultValue Value
	Directives   []*Directive
}

func (vd *VariableDefinition) Pos() int { return vd.Position }
func (vd *VariableDefinition) End() int { return -1 }

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
	Position int
	Name     *Name
}

func (n *NamedType) Pos() int  { return n.Position }
func (n *NamedType) End() int  { return -1 }
func (n *NamedType) typeNode() {}

// ListType
//
// https://spec.graphql.org/draft/#ListType
type ListType struct {
	Position int
	Type     Type
}

func (l *ListType) Pos() int  { return l.Position }
func (l *ListType) End() int  { return -1 }
func (l *ListType) typeNode() {}

// NonNullType
//
// https://spec.graphql.org/draft/#NonNullType
type NonNullType struct {
	Position int
	Type     Type
}

func (n *NonNullType) Pos() int  { return n.Position }
func (n *NonNullType) End() int  { return -1 }
func (n *NonNullType) typeNode() {}

// Name
//
// https://spec.graphql.org/draft/#Name
type Name struct {
	Position int
	Value    string
}

func (n *Name) Pos() int { return n.Position }
func (n *Name) End() int { return -1 }
