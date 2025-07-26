package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"
)

const baseStructTemplate = `package {{.PackageName}}

import (
	"context"
	"time"

	"{{.DBPath}}/model"
	"{{.DBPath}}/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/carsonkrueger/jetkit"
)

type {{.StructName}} struct{}

func (self *{{.StructName}}) InsertCols() postgres.ColumnList {
	return table.{{.Table}}.AllColumns.Except(
{{.InsertCols}}
	)
}

func (self *{{.StructName}}) UpdateCols() postgres.ColumnList {
	return table.{{.Table}}.AllColumns.Except(
{{.UpdateCols}}
	)
}

func (self *{{.StructName}}) AllCols() postgres.ColumnList {
	return table.{{.Table}}.AllColumns
}

func (self *{{.StructName}}) OnConflictCols() postgres.ColumnList {
	return []postgres.Column{}
}

func (self *{{.StructName}}) UpdateOnConflictCols() []postgres.ColumnAssigment {
	return []postgres.ColumnAssigment{}
}

func (self *{{.StructName}}) PKMatch(pk int64) postgres.BoolExpression {
	return {{.PKMatch}}
}

func (self *{{.StructName}}) GetUpdatedAt(row *model.{{.Table}}) *time.Time {
	return {{.UpdatedAt}}
}
`

type TemplateData struct {
	PackageName string
	DBPath      string
	Table       string
	StructName  string
	PKMatch     string
	InsertCols  string
	UpdateCols  string
	UpdatedAt   string
	OutputPath  string
}

type PrimaryKey struct {
	Name    string
	JetType string
}

func main() {
	goFile := os.Getenv("GOFILE")
	goPackage := os.Getenv("GOPACKAGE")
	dbPath := os.Getenv("JET_DB_IMPORT")
	structPrefix := os.Getenv("JETKIT_")
	goFile = "/home/carson/Repos/jetkit/main/main.go" // remove me
	goPackage = "main"                                // remove me
	dbPath = "github.com/jetkit/database"

	var (
		pk                 = flag.String("pk", "", "Primary key column names as key:value pair (e.g. id:string,user_id:integer)")
		schema             = flag.String("schema", "", "Schema name (e.g. public)")
		table              = flag.String("table", "", "Struct/Table name")
		excludedInsertCols = flag.String("excl_insert", "", "Comma-separated column names you want excluded upon insertion")
		excludedUpdateCols = flag.String("excl_update", "", "Comma-separated column names you want excluded upon update")
		updateCol          = flag.String("updated_col", "UpdatedAt", "Column name for UpdatedAt")
		help               = flag.Bool("help", false, "Show help")
		h                  = flag.Bool("h", false, "Show help")
	)
	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	if *pk == "" {
		fmt.Println("Primary key column names are required")
		flag.Usage()
		os.Exit(1)
	}
	if *table == "" {
		fmt.Println("Table name is required")
		flag.Usage()
		os.Exit(1)
	}
	if *schema == "" {
		fmt.Println("Schema name is required")
		flag.Usage()
		os.Exit(1)
	}
	if dbPath == "" {
		fmt.Println("JET_DB_IMPORT env variable is required")
		flag.Usage()
		os.Exit(1)
	}
	if structPrefix == "" {
		structPrefix = "Base"
	}

	dbPath += "/" + *schema
	idx := strings.LastIndex(goFile, "/")
	outputFile := goFile[:idx] + "/" + *table + "_jetkit.go"
	updatedAtStr := ""
	if *updateCol == "" || *updateCol == "nil" {
		updatedAtStr = "nil"
	} else {
		updatedAtStr = "row." + toPascalCase(*updateCol)
	}

	data := TemplateData{
		PackageName: goPackage,
		DBPath:      dbPath,
		Table:       toPascalCase(*table),
		StructName:  toPascalCase(structPrefix) + toPascalCase(*table),
		PKMatch:     buildPKMatch(*table, *pk),
		InsertCols:  formatCols(*excludedInsertCols),
		UpdateCols:  formatCols(*excludedUpdateCols),
		UpdatedAt:   updatedAtStr,
		OutputPath:  outputFile,
	}

	tmpl, err := template.New("BaseStruct").Parse(baseStructTemplate)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		panic(err)
	}

	fmt.Printf("Generated: %s\n", outputFile)
}

func buildPKMatch(table, pkStr string) string {
	pks, err := parsePrimaryKeys(pkStr)
	if err != nil {
		panic(err)
	}
	if len(pks) == 0 {
		panic("No primary keys found")
	}
	builder := strings.Builder{}
	useDotNotation := len(pks) > 1
	for i, pk := range pks {
		if i > 0 {
			builder.WriteString(".\n\t\tOR(")
		}
		dot := ""
		if useDotNotation {
			dot = "." + pk.Name
		}
		_, err = builder.WriteString(fmt.Sprintf("table.%s.%s.EQ(postgres.%s(pk%s))", table, pk.Name, pk.JetType, dot))
		if err != nil {
			panic(err)
		}
		if i > 0 {
			builder.WriteString(")")
		}
	}
	return builder.String()
}

func formatCols(cols string) string {
	if cols == "" {
		return ""
	}
	lines := strings.Builder{}
	for c := range strings.SplitSeq(cols, ",") {
		lines.WriteString("\t\t" + "table." + toPascalCase(c) + ",\n")
	}
	return lines.String()
}

// toPascalCase converts a string in PascalCase, camelCase, or snake_case to PascalCase.
func toPascalCase(input string) string {
	if input == "" {
		return ""
	}

	// Handle snake_case
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == '_' || r == '-' || unicode.IsSpace(r)
	})

	if len(parts) == 1 {
		// Handle camelCase by splitting on case transition
		return capitalizeWords(splitCamelCase(parts[0]))
	}

	// Capitalize each part for PascalCase
	for i := range parts {
		parts[i] = capitalize(parts[i])
	}
	return strings.Join(parts, "")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func splitCamelCase(s string) []string {
	var words []string
	var current []rune

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && (len(current) > 0 && unicode.IsLower(current[len(current)-1])) {
			words = append(words, string(current))
			current = []rune{}
		}
		current = append(current, r)
	}
	words = append(words, string(current))
	return words
}

func capitalizeWords(parts []string) string {
	for i, p := range parts {
		parts[i] = capitalize(p)
	}
	return strings.Join(parts, "")
}

func parsePrimaryKeys(pks string) ([]PrimaryKey, error) {
	keys := []PrimaryKey{}
	for pk := range strings.SplitSeq(pks, ",") {
		parts := strings.Split(pk, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid primary key format, expecting key:value\n %s", pk)
		}
		typ, err := toJetType(parts[1])
		if err != nil {
			return nil, err
		}
		keys = append(keys, PrimaryKey{
			Name:    toPascalCase(parts[0]),
			JetType: typ,
		})
	}
	return keys, nil
}

// toJetType maps PostgreSQL types or Go-like types to Go types
func toJetType(typ string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(typ))

	switch t {
	// PostgreSQL types
	case "text", "varchar", "char", "character varying", "uuid", "inet", "citext", "string", "str":
		return "String", nil
	case "bigint", "int8", "integer", "int", "int4", "smallint", "int2", "int16":
		return "Int", nil
	case "boolean", "bool":
		return "Bool", nil
	case "real", "float4", "double precision", "float8", "float", "float64":
		return "Float", nil
	case "numeric", "decimal", "money":
		return "String", nil
	case "date", "timestamp", "timestamp without time zone", "timestamp with time zone", "timestamptz":
		return "Time", nil
	default:
		return "", fmt.Errorf("unsupported type: %s", typ)
	}
}
