package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/carsonkrueger/jetkit/internal/snaker"
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
	dbPath = "github.com/jetkit/database"             // remove me

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
		updatedAtStr = "row." + snaker.SnakeToCamel(*updateCol, true)
	}

	data := TemplateData{
		PackageName: goPackage,
		DBPath:      dbPath,
		Table:       snaker.SnakeToCamel(*table, true),
		StructName:  snaker.SnakeToCamel(structPrefix, true) + snaker.SnakeToCamel(*table, true),
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
		lines.WriteString("\t\t" + "table." + snaker.SnakeToCamel(c, true) + ",\n")
	}
	return lines.String()
}

func parsePrimaryKeys(pks string) ([]PrimaryKey, error) {
	keys := []PrimaryKey{}
	for pk := range strings.SplitSeq(pks, ",") {
		parts := strings.Split(pk, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("%s invalid primary key format, expecting key:value", pk)
		}
		typ, err := toJetType(parts[1])
		if err != nil {
			return nil, err
		}
		keys = append(keys, PrimaryKey{
			Name:    snaker.SnakeToCamel(parts[0], true),
			JetType: typ,
		})
	}
	return keys, nil
}

// toJetType maps PostgreSQL types or Go-like types to Go types
func toJetType(typ string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(typ))

	switch t {
	case "text", "varchar", "char", "character varying", "uuid", "inet", "citext", "string", "str":
		return "String", nil
	case "bigint", "int8", "integer", "int", "int4", "smallint", "int2", "int16":
		return "Int", nil
	case "boolean", "bool":
		return "Bool", nil
	case "real", "float4", "double precision", "float8", "float", "float64", "double":
		return "Float", nil
	case "numeric", "decimal", "money":
		return "String", nil
	case "date", "timestamp", "timestamp without time zone", "timestamp with time zone", "timestamptz", "datetime":
		return "Time", nil
	default:
		return "", fmt.Errorf("unsupported type: %s", typ)
	}
}
