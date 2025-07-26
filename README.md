
# 🛩️ jetkit

**`jetkit`** is a lightweight extension package for [go-jet](https://github.com/go-jet/jet) providing **easy, type-safe, reusable DAO utilities** for Go projects using Jet with PostgreSQL.

It standardizes **CRUD operations** while retaining Jet’s full type safety and composability.

---

## 🚀 Features

✅ Clean, type-safe **DAO interface** pattern
✅ Easy `Insert`, `Upsert`, `Update`, `Delete`, `GetOne`, `GetMany`, `Index`
✅ Integrates seamlessly with `github.com/go-jet/jet/v2/postgres`
✅ Supports `ON CONFLICT` upserts cleanly
✅ Keeps `updated_at` and primary key matching consistent
✅ Simple, lightweight, zero reflection, zero-cost abstraction

---

## 📦 Installation

```bash
go get github.com/yourname/jetkit
```

## ⚡ Example Usage

Below is a **full practical example** showing how to use `jetkit` in your project.

---

### Generate your models using [jet](https://github.com/go-jet/jet)'s codegen tools

```base
jet -dsn="postgres://db_user:password@localhost:5432/db_name?sslmode=disable" -schema=public -path=./gen
```

### Then implement the interfaces required by jetkit

```go
type UserDAO struct{} // My database access object

func (dao *UserDAO) Table() context.PostgresTable {
	return table.Users
}

func (dao *UserDAO) InsertCols() postgres.ColumnList {
	return table.Users.AllColumns.Except(
		table.Users.ID,
		table.Users.CreatedAt,
		table.Users.UpdatedAt,
	)
}

func (dao *UserDAO) UpdateCols() postgres.ColumnList {
	return table.Users.AllColumns.Except(
		table.Users.ID,
		table.Users.CreatedAt,
	)
}

func (dao *UserDAO) AllCols() postgres.ColumnList {
	return table.Users.AllColumns
}

func (dao *UserDAO) OnConflictCols() postgres.ColumnList {
	return []postgres.Column{table.Users.Name}
}

func (dao *UserDAO) UpdateOnConflictCols() []postgres.ColumnAssigment {
	return []postgres.ColumnAssigment{
		table.Users.Name.SET(table.Users.EXCLUDED.Name),
	}
}

func (dao *UserDAO) PKMatch(pk int64) postgres.BoolExpression {
	return table.Users.ID.EQ(postgres.Int(pk))
}

func (dao *UserDAO) GetUpdatedAt(row *model.Users) *time.Time {
	return row.UpdatedAt
}
```

### Finally use the base queries object

```go
var db *sql.DB
ctx = jetkit.WithDB(context.Background(), db) // base queries object requires db in the ctx
userDAO := UserDAO{}
base := jetkit.NewBaseQueries(&userDAO)
user := model.Users{}
base.Insert(ctx, &user)
```

#### OR call the function directly if you don't like the base queries object

```go
var db *sql.DB
userDAO := UserDAO{}
user := model.Users{}
jetkit.Insert(ctx, &userDAO, &user, db)
```
