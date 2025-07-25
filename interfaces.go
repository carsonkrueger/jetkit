package jetkit

import (
	gctx "context"
	"time"

	"github.com/go-jet/jet/v2/postgres"
)

type GetBaseCols interface {
	InsertCols() postgres.ColumnList
	UpdateCols() postgres.ColumnList
	AllCols() postgres.ColumnList
}

type GetConflictCols interface {
	OnConflictCols() postgres.ColumnList
	UpdateOnConflictCols() []postgres.ColumnAssigment
}

type PKMatcher[PK any] interface {
	PKMatch(pk PK) postgres.BoolExpression
}

type GetUpdatedAt[R any] interface {
	GetUpdatedAt(row *R) *time.Time
}

type BaseQueries[PK PrimaryKey, R any] interface {
	Index(ctx gctx.Context, params *SearchParams) ([]*R, error)
	GetOne(ctx gctx.Context, pk PK) (*R, error)
	GetMany(ctx gctx.Context, pks []PK) ([]*R, error)
	Insert(ctx gctx.Context, model *R) error
	InsertMany(ctx gctx.Context, models []*R) error
	Upsert(ctx gctx.Context, model *R) error
	UpsertMany(ctx gctx.Context, models []*R) error
	Update(ctx gctx.Context, model *R, pk PK) error
	Delete(ctx gctx.Context, pk PK) error
}

// Database Access Object
type DAO[PK PrimaryKey, R any] interface {
	postgres.Table
	GetBaseCols
	GetConflictCols
	PKMatcher[PK]
	GetUpdatedAt[R]
}
