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

func prepareDDL(t *testing.T) *parser.DDL {
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
	ddl := prepareDDL(t)
	opts := ConvertOptions{
		PackageName: "dbrow",
		StructName: "User",
	}
	code, err := convertDDLToStructDef(ddl, opts)
	if err != nil {
		t.Fatal(err)
	}
	snaps.MatchSnapshot(t, code)
}
