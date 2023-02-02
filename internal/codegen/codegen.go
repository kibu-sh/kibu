package codegen

import (
	"bytes"
	"fmt"
	"github.com/discernhq/devx/internal/codedef"
	"github.com/discernhq/devx/internal/codegen/templates"
	"github.com/discernhq/devx/internal/parser"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"strings"
)

type GeneratorFunc func(pkgList parser.PackageList, params GenerateParams) (err error)
type Pipeline []GeneratorFunc

type GenerateParams struct {
	Dir       string
	Pipeline  Pipeline
	Patterns  []string
	OutputDir string
}

func Generate(params GenerateParams) (err error) {
	err = os.MkdirAll(params.OutputDir, os.ModePerm)
	if err != nil {
		return
	}

	pkgList, err := parser.ExperimentalParse(params.Dir, params.Patterns...)
	if err != nil {
		return
	}

	for _, generateFunc := range params.Pipeline {
		if err = generateFunc(pkgList, params); err != nil {
			return
		}
	}
	return
}

func GenerateService(pkgList parser.PackageList, params GenerateParams) (err error) {
	for _, pkg := range pkgList {
		for _, svc := range pkg.Services {
			var data bytes.Buffer
			var target = strings.Replace(svc.Position.Filename, ".go", ".gen.go", 1)
			fmt.Println("generating service", target)
			data, err = templates.Service(&codedef.Module{
				Name: pkg.Name,
				Service: codedef.Service{
					Name:        svc.Name,
					Description: "This a service",
					Type:        "",
					Endpoints: lo.MapToSlice(svc.Endpoints, func(key string, endpoint *parser.Endpoint) codedef.Endpoint {
						return codedef.Endpoint{
							Name:    endpoint.Name,
							Path:    endpoint.Path,
							Methods: endpoint.Methods,
							Request: codedef.Type{
								Name: endpoint.Request.Type,
							},
							Response: codedef.Type{
								Name: endpoint.Response.Type,
							},
						}
					}),
				},
			})
			if err != nil {
				return
			}

			err = os.WriteFile(target, data.Bytes(), os.ModePerm)
			if err != nil {
				return
			}
		}
	}
	return
}

func GenerateWorker(pkgList parser.PackageList, params GenerateParams) (err error) {
	for _, pkg := range pkgList {
		for _, wrk := range pkg.Workers {
			var data bytes.Buffer
			var target = strings.Replace(wrk.Position.Filename, ".go", ".gen.go", 1)
			fmt.Println("generating worker", target)
			data, err = templates.Worker(&codedef.Module{
				Name: pkg.Name,
				Worker: codedef.Worker{
					Name:      wrk.Name,
					Type:      wrk.Type,
					TaskQueue: wrk.TaskQueue,
					Methods: lo.MapToSlice(wrk.Methods, func(key string, method *parser.Method) codedef.Method {
						return codedef.Method{
							Name:        method.Name,
							Description: "",
							Request: codedef.Type{
								Name: method.Request.Name,
								Type: method.Request.Type,
							},
							Response: codedef.Type{
								Name: method.Response.Name,
								Type: method.Response.Type,
							},
						}
					}),
				},
			})
			if err != nil {
				return
			}

			err = os.WriteFile(target, data.Bytes(), os.ModePerm)
			if err != nil {
				return
			}
		}
	}
	return
}

func GenerateHTTPHandlerFactoryContainer(pkgList parser.PackageList, params GenerateParams) (err error) {
	var outFile = filepath.Join(params.OutputDir, "http_handler_factories.gen.go")
	var factory = new(codedef.HTTPHandlerFactoryContainer)

	for _, pkg := range pkgList {
		for _, service := range pkg.Services {
			factory.Imports = append(factory.Imports, pkg.GoPackage.PkgPath)
			factory.Factories = append(factory.Factories, codedef.HTTPHandlerFactory{
				Module: pkg.Name,
				Name:   service.Name,
			})
		}
	}

	if len(factory.Factories) == 0 {
		return
	}

	fmt.Println("generating http handler factory container", outFile)

	data, err := templates.HttpHandlerFactoryContainer(factory)
	if err != nil {
		return
	}

	err = os.WriteFile(outFile, data.Bytes(), os.ModePerm)
	return
}

func DefaultPipeline() Pipeline {
	return Pipeline{
		GenerateService,
		GenerateWorker,
		GenerateHTTPHandlerFactoryContainer,
	}
}
