package main

type Options struct {
	SchemaPath  string
	Table       string
	OutFilePath string

	ConvertOptions
}

type ConvertOptions struct {
	PackageName string
	StructName  string
}
