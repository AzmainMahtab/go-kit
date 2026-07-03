package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// txKey is the context key used to store the current transaction.
type txKey struct{}

// Executor is the interface repositories need from a SQL executor.
// Both *sqlx.DB and *sqlx.Tx satisfy it.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
}

// TxManager coordinates database transactions.
//
// Use cases receive a context; repositories check the context for an active
// transaction. If one exists they use it, otherwise they use the regular DB.
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager creates a TxManager.
func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// WithinTransaction executes fn inside a transaction.
func (tm *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v; rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Executor returns either the active transaction from ctx or the base DB.
func (tm *TxManager) Executor(ctx context.Context) Executor {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return tm.db
}

// GetDB returns the underlying *sqlx.DB.
func (tm *TxManager) GetDB() *sqlx.DB {
	return tm.db
}
