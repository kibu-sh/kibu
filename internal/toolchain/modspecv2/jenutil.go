package modspecv2

import "github.com/dave/jennifer/jen"

func MultiLineParen() jen.Options {
	return jen.Options{
		Open:      "(",
		Close:     ")",
		Separator: ",",
		Multi:     true,
	}
}

func MultiLineCurly() jen.Options {
	return jen.Options{
		Open:      "{",
		Close:     "}",
		Separator: ",",
		Multi:     true,
	}
}
