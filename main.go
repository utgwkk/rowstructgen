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
	"golang.org/x/tools/imports"
)

var schemaPath string
var targetTable string
var packageName string
var structName string

func init() {
	flag.StringVar(&schemaPath, "schema", "-", "path to schema file")
	flag.StringVar(&targetTable, "table", "", "table name to generate struct definition")
	flag.StringVar(&packageName, "package", "", "package name to generate struct definition")
	flag.StringVar(&structName, "struct", "", "struct name to generate definition")
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

func extractTableDefinition(ddls []database.DDLStatement, tableName string) (*parser.DDL, error) {
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
		return "*"+goType
	}
	return goType
}

func convertDDLToStructDef(ddl *parser.DDL, packageName, structName string) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	buf.WriteString(fmt.Sprintf("type %s struct {\n", structName))

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

	if targetTable == "" {
		log.Fatal("table name not set")
	}

	if structName == "" {
		log.Fatal("struct name not set")
	}

	schema, err := readSchema(schemaPath)
	if err != nil {
		log.Fatal(err)
	}

	mysqlParser := database.NewParser(parser.ParserModeMysql)
	ddls, err := mysqlParser.Parse(schema)
	if err != nil {
		log.Fatal(err)
	}

	ddl, err := extractTableDefinition(ddls, targetTable)
	if err != nil {
		log.Fatal(err)
	}

	structDef, err := convertDDLToStructDef(ddl, packageName, structName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(structDef)
}