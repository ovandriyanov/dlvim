package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

func main() {
	args := parseCommandLineArguments()

	packageAST := parseAST(args.inputDirectoryName)

	rpcMethods := listRPCMethods(packageAST, args.rpcHandlerTypeName)
	if len(rpcMethods) == 0 {
		exit(fmt.Errorf("No RPC methods found for %s in %s", args.rpcHandlerTypeName, args.inputDirectoryName))
	}

	outputFile, err := os.Create(args.outputFileName)
	if err != nil {
		exit(fmt.Errorf("Cannot open %s: %w\n", args.outputFileName, err))
	}
	defer outputFile.Close()

	writeOutput(packageAST.Name, args.rpcHandlerTypeName, rpcMethods, outputFile)
}

func writeOutput(packageName string, rpcHandlerTypeName string, rpcMethods []rpcMethod, writer io.Writer) {
	fmt.Fprintf(writer, "// Generated with generate_logging_rpc_handler. Do not edit!\n")
	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "package %s\n", packageName)
	fmt.Fprintf(writer, "\n")
	writeImports(rpcMethods, writer)
	fmt.Fprintf(writer, "type Logging%s struct {\n", rpcHandlerTypeName)
	fmt.Fprintf(writer, "\tserverName string\n")
	fmt.Fprintf(writer, "\twrappedHandler *%s\n", rpcHandlerTypeName)
	fmt.Fprintf(writer, "}\n")
	fmt.Fprintf(writer, "\n")

	for _, rpcMethod := range rpcMethods {
		fmt.Fprintf(writer, "func (h *Logging%s) %s(request %s, response %s) error {\n", rpcHandlerTypeName, rpcMethod.name, rpcMethod.requestType.expression, rpcMethod.responseType.expression)
		fmt.Fprintf(writer, "\tmarshaledRequest, _ := json.Marshal(request)\n")
		fmt.Fprintf(writer, "\tlog.Printf(\"%%s: <-- %s %%s\\n\", h.serverName, string(marshaledRequest))\n", rpcMethod.name)
		fmt.Fprintf(writer, "\terr := h.wrappedHandler.%s(request, response)\n", rpcMethod.name)
		fmt.Fprintf(writer, "\tif err != nil {\n")
		fmt.Fprintf(writer, "\t\tlog.Printf(\"%%s: --> %s error %%v\\n\", h.serverName, err)\n", rpcMethod.name)
		fmt.Fprintf(writer, "\t\treturn err\n")
		fmt.Fprintf(writer, "\t}\n")
		fmt.Fprintf(writer, "\tmarshaledResponse, _ := json.Marshal(response)\n")
		fmt.Fprintf(writer, "\tlog.Printf(\"%%s: --> %s %%s\\n\", h.serverName, string(marshaledResponse))\n", rpcMethod.name)
		fmt.Fprintf(writer, "\treturn nil\n")
		fmt.Fprintf(writer, "}\n")
		fmt.Fprintf(writer, "\n")
	}

	fmt.Fprintf(writer, "func NewLogging%s(wrappedHandler *%s, serverName string) *Logging%s {\n", rpcHandlerTypeName, rpcHandlerTypeName, rpcHandlerTypeName)
	fmt.Fprintf(writer, "\treturn &Logging%s{\n", rpcHandlerTypeName)
	fmt.Fprintf(writer, "\t\tserverName: serverName,\n")
	fmt.Fprintf(writer, "\t\twrappedHandler: wrappedHandler,\n")
	fmt.Fprintf(writer, "\t}\n")
	fmt.Fprintf(writer, "}\n")
}

func writeImports(rpcMethods []rpcMethod, writer io.Writer) {
	importByName := map[string]*ast.ImportSpec{
		"json": {Name: &ast.Ident{Name: "json"}, Path: &ast.BasicLit{Value: `"encoding/json"`}},
		"log":  {Name: &ast.Ident{Name: "log"}, Path: &ast.BasicLit{Value: `"log"`}},
	}
	for _, method := range rpcMethods {
		if requestImport := method.requestType.importSpec; requestImport != nil {
			importByName[requestImport.Name.String()] = requestImport
		}
		if responseImport := method.responseType.importSpec; responseImport != nil {
			importByName[responseImport.Name.String()] = responseImport
		}
	}

	sortedImports := make([]*ast.ImportSpec, 0, len(importByName))
	for _, oneImport := range importByName {
		sortedImports = append(sortedImports, oneImport)
	}
	sort.Slice(sortedImports, func(i, j int) bool { return sortedImports[i].Path.Value < sortedImports[j].Path.Value })

	fmt.Fprintf(writer, "import (\n")
	for _, oneImport := range sortedImports {
		fmt.Fprintf(writer, "\t%s %s\n", oneImport.Name.String(), oneImport.Path.Value)
	}
	fmt.Fprintf(writer, ")\n")
	fmt.Fprintf(writer, "\n")
}

func parseAST(directoryName string) *ast.Package {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, directoryName, nil, parser.AllErrors)
	if err != nil {
		exit(fmt.Errorf("cannot parse %s: %w", directoryName, err))
	}
	if len(packages) == 0 {
		exit(fmt.Errorf("No packages found in %s", directoryName))
	}
	if len(packages) > 1 {
		exit(fmt.Errorf("Parsing more than one package in a directory is not supported (found %d packages)", len(packages)))
	}

	for _, pkg := range packages {
		return pkg
	}
	return nil
}

type typeDescription struct {
	expression string
	importSpec *ast.ImportSpec
}

type rpcMethod struct {
	name         string
	requestType  typeDescription
	responseType typeDescription
}

func listRPCMethods(packageAST *ast.Package, rpcHandlerTypeName string) []rpcMethod {
	var methods []rpcMethod
	for _, astFile := range packageAST.Files {
		fileImports := map[string]*ast.ImportSpec{}
		for _, importSpec := range astFile.Imports {
			fileImports[importSpec.Name.String()] = importSpec
		}
		for _, declaration := range astFile.Decls {
			functionDeclaration, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			receiverDeclaration := functionDeclaration.Recv
			if receiverDeclaration == nil || len(receiverDeclaration.List) == 0 {
				continue
			}
			receiverField := receiverDeclaration.List[0]
			receiverExpression, ok := receiverField.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			receiverIdentifier, ok := receiverExpression.X.(*ast.Ident)
			if !ok {
				continue
			}
			if receiverIdentifier.String() != rpcHandlerTypeName {
				continue
			}

			if functionDeclaration.Name == nil || !functionDeclaration.Name.IsExported() {
				continue
			}
			if functionDeclaration.Type == nil || functionDeclaration.Type.Params == nil || functionDeclaration.Type.Results == nil {
				continue
			}
			paramsList := functionDeclaration.Type.Params.List
			if paramsList == nil || len(paramsList) != 2 {
				continue
			}
			requestParam := paramsList[0]
			if len(requestParam.Names) != 1 {
				continue
			}
			if requestParam.Type == nil {
				continue
			}
			var requestImportSpec *ast.ImportSpec
			if starExpr, ok := requestParam.Type.(*ast.StarExpr); ok {
				if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
					requestImportSpec = fileImports[types.ExprString(selectorExpr.X)]
				}
			}
			requestTypeDescription := typeDescription{
				expression: types.ExprString(requestParam.Type),
				importSpec: requestImportSpec,
			}

			responseParam := paramsList[1]
			if len(responseParam.Names) != 1 {
				continue
			}
			if responseParam.Type == nil {
				continue
			}
			var responseImportSpec *ast.ImportSpec
			if starExpr, ok := responseParam.Type.(*ast.StarExpr); ok {
				if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
					responseImportSpec = fileImports[types.ExprString(selectorExpr.X)]
				}
			}
			responseTypeDescription := typeDescription{
				expression: types.ExprString(responseParam.Type),
				importSpec: responseImportSpec,
			}

			resultsList := functionDeclaration.Type.Results.List
			if resultsList == nil || len(resultsList) != 1 {
				continue
			}
			resultTypeExpr, ok := resultsList[0].Type.(*ast.Ident)
			if !ok {
				continue
			}
			if resultTypeExpr.Name != "error" {
				continue
			}

			methods = append(methods, rpcMethod{
				name:         functionDeclaration.Name.String(),
				requestType:  requestTypeDescription,
				responseType: responseTypeDescription,
			})
		}
	}
	return methods
}

type arguments struct {
	inputDirectoryName string
	outputFileName     string
	rpcHandlerTypeName string
}

func parseCommandLineArguments() arguments {
	inputDirectoryName := flag.String("input-dir", "", "Directory with .go files to generate output from")
	outputFileName := flag.String("output-file", "", "Name of the file to write the output into")
	rpcHandlerTypeName := flag.String("rpc-handler-type-name", "", "Name of the RPC handler type to wrap")
	flag.Parse()
	if len(*inputDirectoryName) == 0 {
		exit("No --input-dir flag provided")
	}
	if len(*outputFileName) == 0 {
		exit("No --output-file flag provided")
	}
	if len(*rpcHandlerTypeName) == 0 {
		exit("No --rpc-handler-type-name flag provided")
	}
	return arguments{
		inputDirectoryName: *inputDirectoryName,
		outputFileName:     *outputFileName,
		rpcHandlerTypeName: *rpcHandlerTypeName,
	}
}

func exit(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
