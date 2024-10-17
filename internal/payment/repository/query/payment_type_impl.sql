-- name: CreatePaymentType :one
INSERT INTO payment_type (
    ptname
) VALUES  (
    $1
) RETURNING *;

-- name: GetPaymentTypeByName :one
SELECT * FROM payment_type WHERE ptname = $1 LIMIT 1;