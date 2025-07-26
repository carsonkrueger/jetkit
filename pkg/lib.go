package jetkit

import (
	gctx "context"
	"errors"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

var (
	ErrNilParams = errors.New("cannot use nil params in query")
)

// params are optional and can be used to filter, sort, paginate, and limit the results of a query.
func Index[PK PrimaryKey, R any, D DAO[PK, R]](ctx gctx.Context, dao D, params *SearchParams, db qrm.Queryable) ([]*R, error) {
	query := dao.SELECT(dao.AllCols())
	if params != nil {
		if params.Where != nil {
			query = query.WHERE(params.Where)
		}
		if len(params.OrderBy) > 0 {
			query = query.ORDER_BY(params.OrderBy...)
		}
		if params.Offset != nil {
			query = query.OFFSET(*params.Offset)
		}
		if params.Limit != nil {
			query = query.LIMIT(*params.Limit)
		}
	}
	models := []*R{}
	if err := query.QueryContext(ctx, db, &models); err != nil {
		return nil, err
	}
	return models, nil
}

func GetOne[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, pk PK, db Q) (*R, error) {
	var model R
	if err := dao.
		SELECT(dao.AllCols()).
		WHERE(dao.PKMatch(pk)).
		LIMIT(1).
		QueryContext(ctx, db, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

func GetMany[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, pks []PK, db Q) ([]*R, error) {
	if pks == nil {
		return nil, ErrNilParams
	}
	rows := []*R{}
	if len(pks) == 0 {
		return rows, nil
	}
	where := dao.PKMatch(pks[0])
	for _, pk := range pks[1:] {
		where = where.OR(dao.PKMatch(pk))
	}
	if err := dao.
		SELECT(dao.AllCols()).
		WHERE(where).
		QueryContext(ctx, db, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func Insert[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, row *R, db Q) error {
	if row == nil {
		return ErrNilParams
	}
	return dao.
		INSERT(dao.InsertCols()).
		MODEL(row).
		RETURNING(dao.AllCols()).
		QueryContext(ctx, db, row)
}

func InsertMany[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, rows []*R, db Q) error {
	if rows == nil {
		return ErrNilParams
	}
	return dao.
		INSERT(dao.InsertCols()).
		MODELS(rows).
		RETURNING(dao.AllCols()).
		QueryContext(ctx, db, rows)
}

func Upsert[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, model *R, db Q) error {
	if model == nil {
		return ErrNilParams
	}
	up := dao.GetUpdatedAt(model)
	if up != nil {
		*up = time.Now()
	}
	conflictCols := dao.OnConflictCols()
	updateCols := dao.UpdateOnConflictCols()
	query := dao.
		INSERT(dao.UpdateCols()).
		MODEL(model)
	if len(updateCols) > 0 && len(conflictCols) > 0 {
		query = query.
			ON_CONFLICT(conflictCols...).
			DO_UPDATE(postgres.SET(updateCols...))
	}
	return query.
		RETURNING(dao.AllCols()).
		QueryContext(ctx, db, model)
}

func UpsertMany[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, rows []*R, db Q) error {
	if rows == nil {
		return ErrNilParams
	}
	for _, v := range rows {
		up := dao.GetUpdatedAt(v)
		if up != nil {
			*up = time.Now()
		}
	}
	conflictCols := dao.OnConflictCols()
	updateCols := dao.UpdateOnConflictCols()
	query := dao.
		INSERT(dao.UpdateCols()).
		MODELS(rows)
	if len(updateCols) > 0 && len(conflictCols) > 0 {
		query = query.
			ON_CONFLICT(conflictCols...).
			DO_UPDATE(postgres.SET(updateCols...))
	}
	return query.
		RETURNING(dao.AllCols()).
		QueryContext(ctx, db, rows)
}

func Update[PK PrimaryKey, R any, D DAO[PK, R], Q qrm.Queryable](ctx gctx.Context, dao D, model *R, pk PK, db Q) error {
	if model == nil {
		return ErrNilParams
	}
	up := dao.GetUpdatedAt(model)
	if up != nil {
		*up = time.Now()
	}
	return dao.
		UPDATE(dao.UpdateCols()).
		MODEL(model).
		WHERE(dao.PKMatch(pk)).
		RETURNING(dao.AllCols()).
		QueryContext(ctx, db, model)
}

func Delete[PK PrimaryKey, R any, D DAO[PK, R], E qrm.Executable](ctx gctx.Context, dao D, pk PK, db E) error {
	_, err := dao.
		DELETE().
		WHERE(dao.PKMatch(pk)).
		ExecContext(ctx, db)
	return err
}

type baseQueryable[PK PrimaryKey, R any] struct {
	dao DAO[PK, R]
}

// BaseQueries implements a list of easy methods for jet. This requires the DB connection to be available through the context, using jetkit.WithDB()
func NewQueryable[PK PrimaryKey, R any](dao DAO[PK, R]) BaseQueries[PK, R] {
	return &baseQueryable[PK, R]{
		dao,
	}
}

func (q *baseQueryable[PK, R]) Index(ctx gctx.Context, params *SearchParams) ([]*R, error) {
	return Index(ctx, q.dao, params, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) GetOne(ctx gctx.Context, pk PK) (*R, error) {
	return GetOne(ctx, q.dao, pk, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) GetMany(ctx gctx.Context, pks []PK) ([]*R, error) {
	return GetMany(ctx, q.dao, pks, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) Insert(ctx gctx.Context, model *R) error {
	return Insert(ctx, q.dao, model, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) InsertMany(ctx gctx.Context, models []*R) error {
	return InsertMany(ctx, q.dao, models, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) Upsert(ctx gctx.Context, model *R) error {
	return Upsert(ctx, q.dao, model, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) UpsertMany(ctx gctx.Context, models []*R) error {
	return UpsertMany(ctx, q.dao, models, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) Update(ctx gctx.Context, model *R, pk PK) error {
	return Update(ctx, q.dao, model, pk, GetDB(ctx))
}

func (q *baseQueryable[PK, R]) Delete(ctx gctx.Context, pk PK) error {
	return Delete(ctx, q.dao, pk, GetDB(ctx))
}
