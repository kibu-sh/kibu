package devx

#TypeMap: [name=string]: #Type & {Name: name}

#Type: {
	Name: string
	Fields: [field=string]: #Field & {Name: field}
}

#Field: {
	Name:        string
	Type:        string
	Required:    *false | true
	Description: *"" | string
}

#Handler: {
	Request:  #Type
	Response: #Type
}

#Module: {
	Types?:   #TypeMap
	Service?: #Service
	Worker?:  #Worker
}
