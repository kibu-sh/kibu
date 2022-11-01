package devx

#Middleware: {}

#Param: {
	Name:         string
	From:         *"path" | "query" | "header" | "cookie"
	Required:     true | *false
	Description?: string
}

#Method: "get" | "post" | "put" | "delete" | "patch" | "head" |
	"options" | "connect" | "trace" | "all"

#Endpoint: {
	Name:   string
	Path:   string
	Method: #Method
	Tags?: [...string]
	Description?: string
	Middleware?: [...#Middleware]
	Params: [param=string]: #Param & {Name: param}
	Request:  #Type
	Response: #Type
}

#Service: {
	Name: string
	Type: *"private" | "auth" | "public"
	Tags?: [...string]
	Middleware?: [...#Middleware]

	// TODO: validate path expression
	// https://cuelang.org/docs/tutorials/tour/expressions/regexp/
	Endpoints: [path=string]: {
		Middleware?: [...#Middleware]
		[method=#Method]: #Endpoint & {
			Path: path, Method: method
		}
	}
}
