package store

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type queryReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (s *Store) readDB() queryReader {
	if s.reader != nil {
		return s.reader
	}
	return s.Pool
}

// All contributing reads share one database snapshot, including live review decisions.
func (s *Store) readSnapshot(ctx context.Context, read func(*Store) error) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = read(&Store{Pool: s.Pool, reader: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
