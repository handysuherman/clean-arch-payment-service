// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: payment_type_impl.sql

package repository

import (
	"context"
)

const createPaymentType = `-- name: CreatePaymentType :one
INSERT INTO payment_type (
    ptname
) VALUES  (
    $1
) RETURNING ptname
`

func (q *Queries) CreatePaymentType(ctx context.Context, ptname string) (string, error) {
	row := q.db.QueryRow(ctx, createPaymentType, ptname)
	err := row.Scan(&ptname)
	return ptname, err
}

const getPaymentTypeByName = `-- name: GetPaymentTypeByName :one
SELECT ptname FROM payment_type WHERE ptname = $1 LIMIT 1
`

func (q *Queries) GetPaymentTypeByName(ctx context.Context, ptname string) (string, error) {
	row := q.db.QueryRow(ctx, getPaymentTypeByName, ptname)
	err := row.Scan(&ptname)
	return ptname, err
}
