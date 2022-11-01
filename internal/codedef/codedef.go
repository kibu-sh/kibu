package codedef

type Module struct {
	Name   string
	Types  TypeMap
	Worker Worker
}

type Field struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type Type struct {
	Name   string
	Fields map[string]Field
}

type TypeMap map[string]Type

type Worker struct {
	TaskQueue  string
	Workflows  map[string]Workflow
	Activities map[string]Activity
}

type Workflow struct {
	Name        string
	Description string
	Request     Type
	Response    Type
}

type Activity struct {
	Name        string
	Description string
	Request     Type
	Response    Type
}
