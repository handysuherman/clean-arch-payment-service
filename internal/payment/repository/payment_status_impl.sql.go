// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: payment_status_impl.sql

package repository

import (
	"context"
)

const createPaymentStatus = `-- name: CreatePaymentStatus :one
INSERT INTO payment_status (
    psname
) VALUES (
    $1
) RETURNING psname
`

func (q *Queries) CreatePaymentStatus(ctx context.Context, psname string) (string, error) {
	row := q.db.QueryRow(ctx, createPaymentStatus, psname)
	err := row.Scan(&psname)
	return psname, err
}

const getPaymentStatusByName = `-- name: GetPaymentStatusByName :one
SELECT psname FROM payment_status WHERE psname = $1 LIMIT 1
`

func (q *Queries) GetPaymentStatusByName(ctx context.Context, psname string) (string, error) {
	row := q.db.QueryRow(ctx, getPaymentStatusByName, psname)
	err := row.Scan(&psname)
	return psname, err
}