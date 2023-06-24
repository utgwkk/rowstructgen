package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/k0kubun/sqldef"
	"github.com/k0kubun/sqldef/database"
	"github.com/k0kubun/sqldef/parser"
	"github.com/stoewer/go-strcase"
	"github.com/utgwkk/rowstructgen/options"
	"golang.org/x/tools/imports"
)

var schemaPath string
var targetTable string
var packageName string
var structName string
var outFilePath string
var tableNameConst bool

func init() {
	flag.StringVar(&schemaPath, "schema", "-", "path to schema file (default: stdin)")
	flag.StringVar(&targetTable, "table", "", "table name to generate struct definition")
	flag.StringVar(&packageName, "package", "", "package name to generate struct definition")
	flag.StringVar(&structName, "struct", "", "struct name to generate definition")
	flag.StringVar(&outFilePath, "out", "", "path to generate schema definition (default: stdout)")
	flag.BoolVar(&tableNameConst, "table-name-const", false, "generate table name constant variable (default: false)")
}

func readSchema(path string) (string, error) {
	var r io.ReadCloser
	if path == "-" {
		r = io.NopCloser(os.Stdin)
	} else {
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open schema file: %w", err)
		}
		r = f
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read schema file: %w", err)
	}

	return string(b), nil
}

func extractTableDefinition(ddls []database.DDLStatement, opts *options.Options) (*parser.DDL, error) {
	tableName := opts.Table
	for _, ddl := range ddls {
		switch ddl := ddl.Statement.(type) {
		case *parser.DDL:
			if ddl.Action == "create" && ddl.NewName.Name.String() == tableName {
				return ddl, nil
			}
		}
	}

	return nil, fmt.Errorf("table not found: %s", tableName)
}

func mysqlTypeToGoType(col *parser.ColumnDefinition) (gotype string, acceptNull bool) {
	switch strings.ToLower(strings.ToLower(col.Type.Type)) {
	case "binary", "varbinary", "tinyblob", "blob", "mediumblob", "longblob":
		return "[]byte", true
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		return "string", false
	case "enum":
		// TODO
		return "string", false
	case "date", "datetime", "timestamp":
		return "time.Time", false
	case "boolean":
		return "bool", false
	case "tinyint", "smallint", "mediumint", "int", "integer":
		if col.Type.Unsigned {
			return "uint", false
		}
		return "int", false
	case "bigint":
		if col.Type.Unsigned {
			return "uint64", false
		}
		return "int64", false
	case "float", "double":
		return "float64", false
	case "json":
		return "[]byte", true
	}

	return fmt.Sprintf("!!INVALID: %s!!", col.Type.Type), false
}

func columnTypeToGoType(col *parser.ColumnDefinition) string {
	nullable := true
	if col.Type.NotNull != nil && *col.Type.NotNull {
		nullable = false
	}

	goType, acceptNull := mysqlTypeToGoType(col)
	if nullable && !acceptNull {
		return "*" + goType
	}
	return goType
}

func convertDDLToStructDef(ddl *parser.DDL, opts options.ConvertOptions) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("package %s\n\n", opts.PackageName))

	if opts.GenerateTableNameConstant {
		tableNameUpper := strcase.UpperCamelCase(opts.TableName)
		buf.WriteString(fmt.Sprintf("const Table%s = \"%s\"\n\n", tableNameUpper, opts.TableName))
	}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", opts.StructName))

	for _, col := range ddl.TableSpec.Columns {
		fieldName := strcase.UpperCamelCase(col.Name.String())
		goType := columnTypeToGoType(col)
		buf.WriteString(fmt.Sprintf("\t%s %s `db:\"%s\"`\n", fieldName, goType, col.Name.String()))
	}

	buf.WriteString("}\n")

	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

func main() {
	flag.Parse()

	opts, err := options.New(
		schemaPath,
		targetTable,
		outFilePath,
		packageName,
		structName,
		tableNameConst,
	)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := readSchema(opts.SchemaPath)
	if err != nil {
		log.Fatal(err)
	}

	mysqlParser := database.NewParser(parser.ParserModeMysql)
	ddls, err := mysqlParser.Parse(schema)
	if err != nil {
		log.Fatal(err)
	}

	ddl, err := extractTableDefinition(ddls, opts)
	if err != nil {
		log.Fatal(err)
	}

	structDef, err := convertDDLToStructDef(ddl, opts.ConvertOptions)
	if err != nil {
		log.Fatal(err)
	}

	if opts.OutFilePath == "" {
		fmt.Print(structDef)
	} else {
		f, err := os.Create(opts.OutFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if _, err := f.WriteString(structDef); err != nil {
			log.Fatal(err)
		}
	}
}
