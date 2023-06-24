package options

import (
	"errors"

	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
)

type Options struct {
	SchemaPath  string
	Table       string
	OutFilePath string

	ConvertOptions
}

type ConvertOptions struct {
	PackageName               string
	StructName                string
	TableName                 string
	GenerateTableNameConstant bool
}

func New(
	schemaPath string,
	targetTable string,
	outFilePath string,
	packageName string,
	structName string,
	tableNameConst bool,
) (*Options, error) {
	opts := &Options{
		SchemaPath:  schemaPath,
		Table:       targetTable,
		OutFilePath: outFilePath,

		ConvertOptions: ConvertOptions{
			PackageName:               packageName,
			StructName:                structName,
			TableName:                 targetTable,
			GenerateTableNameConstant: tableNameConst,
		},
	}

	if opts.ConvertOptions.StructName == "" {
		opts.ConvertOptions.StructName = guessStructNameFromTable(opts.Table)
	}

	if opts.Table == "" {
		return nil, errors.New("table name not set")
	}

	return opts, nil
}

func guessStructNameFromTable(tableName string) string {
	pluralized := pluralize.NewClient().Singular(tableName)
	return strcase.UpperCamelCase(pluralized)
}
