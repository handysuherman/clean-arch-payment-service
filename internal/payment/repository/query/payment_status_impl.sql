-- name: CreatePaymentStatus :one
INSERT INTO payment_status (
    psname
) VALUES (
    $1
) RETURNING *;

-- name: GetPaymentStatusByName :one
SELECT * FROM payment_status WHERE psname = $1 LIMIT 1;