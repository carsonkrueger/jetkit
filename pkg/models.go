package jetkit

import "github.com/go-jet/jet/v2/postgres"

type PrimaryKey any

type SearchParams struct {
	Where   postgres.BoolExpression
	OrderBy []postgres.OrderByClause
	Offset  *int64
	Limit   *int64
}
