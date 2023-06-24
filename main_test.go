package main

import (
	_ "embed"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/k0kubun/sqldef/database"
	"github.com/k0kubun/sqldef/parser"
	"github.com/utgwkk/rowstructgen/options"
)

//go:embed testdata/schema.sql
var schema string

func prepareDDL(t *testing.T, tableName string) *parser.DDL {
	mysqlParser := database.NewParser(parser.ParserModeMysql)
	ddls, err := mysqlParser.Parse(schema)
	if err != nil {
		t.Fatal(err)
	}

	ddl, err := extractTableDefinition(ddls, &options.Options{
		Table: "users",
	})
	if err != nil {
		t.Fatal(err)
	}

	return ddl
}

func TestConvertDDLToStructDef(t *testing.T) {
	testcases := []struct {
		name string
		opts options.ConvertOptions
	}{
		{
			name: "default",
			opts: options.ConvertOptions{
				PackageName:               "dbrow",
				TableName:                 "users",
				StructName:                "User",
				GenerateTableNameConstant: false,
			},
		},
		{
			name: "with table name constants",
			opts: options.ConvertOptions{
				PackageName:               "dbrow",
				TableName:                 "users",
				StructName:                "User",
				GenerateTableNameConstant: true,
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ddl := prepareDDL(t, tc.opts.TableName)
			code, err := convertDDLToStructDef(ddl, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			snaps.MatchSnapshot(t, code)
		})
	}
}
