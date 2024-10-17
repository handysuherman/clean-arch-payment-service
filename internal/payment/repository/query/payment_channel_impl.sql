-- name: CreatePaymentChannel :one
INSERT INTO payment_channel (
    uid,
    pcname,
    pc_type,
    min_amount,
    max_amount,
    tax,
    is_tax_percentage
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPaymentChannelByID :one
SELECT * FROM payment_channel WHERE uid = $1 LIMIT 1;

-- name: GetPaymentChannelByName :one
SELECT * FROM payment_channel WHERE pcname = $1 LIMIT 1;

-- name: GetAvailablePaymentChannels :many
SELECT * FROM payment_channel WHERE  $1 > min_amount  AND $1 < max_amount AND is_active = true AND is_available = true;

-- name: GetAvailablePaymentChannel :one
SELECT * FROM payment_channel WHERE pc_type = $1 AND pcname = $2 AND $3 > min_amount AND $3 < max_amount AND is_active = true AND is_available = true LIMIT 1;