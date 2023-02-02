package codedef

type HTTPHandlerFactory struct {
	Module string
	Name   string
}

type HTTPHandlerFactoryContainer struct {
	Imports   []string
	Factories []HTTPHandlerFactory
}

type Module struct {
	Name    string
	Worker  Worker
	Service Service
}

type Field struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type Type struct {
	Name   string
	Fields []Field
	Type   string
}

type Worker struct {
	Name      string
	Type      string
	TaskQueue string
	Methods   []Method
}

type Endpoint struct {
	Name     string
	Path     string
	Methods  []string
	Request  Type
	Response Type
}

type Method struct {
	Name        string
	Description string
	Request     Type
	Response    Type
}

type Service struct {
	Name        string
	Description string
	Type        string
	Endpoints   []Endpoint
}
