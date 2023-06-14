package main

import (
	_ "embed"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/k0kubun/sqldef/database"
	"github.com/k0kubun/sqldef/parser"
)

//go:embed testdata/schema.sql
var schema string

func prepareDDL(t *testing.T, tableName string) *parser.DDL {
	mysqlParser := database.NewParser(parser.ParserModeMysql)
	ddls, err := mysqlParser.Parse(schema)
	if err != nil {
		t.Fatal(err)
	}

	ddl, err := extractTableDefinition(ddls, Options{
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
		opts ConvertOptions
	}{
		{
			name: "default",
			opts: ConvertOptions{
				PackageName: "dbrow",
				TableName:   "users",
				StructName:  "User",
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
