package devx

#Middleware: {}

#Param: {
	Name:         string
	From:         *"path" | "query" | "header" | "cookie"
	Required:     true | *false
	Description?: string
}

#Method: {
	Name: string
	Tags?: [...string]
	Description?: string
	Middleware?: [...#Middleware]
	Params: [param=string]: #Param & {Name: param}

	#Handler

	Transports: {
		HTTP: #HTTP
	}
}

#HTTP: {
	Method: "get" | *"post" | "put" | "delete" | "patch"
	Path:   string
}

#Service: {
	Name: string
	Type: "public" | "auth" | *"private"
	Tags?: [...string]
	Middleware?: [...#Middleware]

	// TODO: validate path expression
	// https://cuelang.org/docs/tutorials/tour/expressions/regexp/
	_ServiceName: Name

	Methods: [name=string]: #Method & {
		Name:        name
		_MethodName: name

		Transports: HTTP: {
			Path: "\(_ServiceName).\(_MethodName)"
		}
	}
}
