package options

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

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

	s, err := os.Stat(outFilePath)
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		opts.OutFilePath = guessOutFilePathFromTable(opts.OutFilePath, opts.Table)

		if opts.PackageName == "" {
			opts.PackageName = guessPackageNameFromOutFilePath(opts.OutFilePath)
		}
	}

	return opts, nil
}

func guessStructNameFromTable(tableName string) string {
	pluralized := pluralize.NewClient().Singular(tableName)
	return strcase.UpperCamelCase(pluralized)
}

func guessOutFilePathFromTable(basePath, tableName string) string {
	singularTableName := pluralize.NewClient().Singular(tableName)
	return path.Join(basePath, singularTableName+".go")
}

func guessPackageNameFromOutFilePath(outFilePath string) string {
	normalizedDir := filepath.Dir(filepath.ToSlash(outFilePath))
	xs := strings.Split(normalizedDir, "/")
	candidate := xs[len(xs)-1]
	if candidate == "." || candidate == ".." {
		return "main"
	}
	return candidate
}
