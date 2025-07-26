package jetkit

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-jet/jet/v2/qrm"
)

var (
	ErrUnsupportedDBType = errors.New("unsupported database type")
	ErrTxAlreadyStarted  = errors.New("transaction already started")
)

type dbKeyType struct{}

var dbKey any = dbKeyType{} // default key

// Override the default key used to store the database connection in the context.
func SetDBKey(key any) {
	dbKey = key
}

// Pass in *sql.DB
func WithDB(ctx context.Context, db qrm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

func GetDB(ctx context.Context) qrm.DB {
	return ctx.Value(dbKey).(qrm.DB)
}

// Returns a new context that contains the transaction. Caller must Rollback and Commit manually. The new tx ctx cannot and should NOT be used after a rollback or commit.
//
// Suggested usage:
//
// txCtx, tx, err := BeginTx(ctx)
//
//	if err != nil {
//	    return err
//	}
//
// defer tx.Rollback()
//
// // Perform database operations using new txCtx
//
// tx.Commit()
func BeginTx(ctx context.Context) (context.Context, *sql.Tx, error) {
	switch db := GetDB(ctx).(type) {
	case *sql.DB:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ctx, nil, err
		}
		ctx = WithDB(ctx, tx)
		return ctx, tx, nil
	case *sql.Tx:
		return ctx, db, ErrTxAlreadyStarted
	default:
		return ctx, nil, ErrUnsupportedDBType
	}
}
