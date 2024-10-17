-- name: CreatePaymentMethod :one
INSERT INTO payment_method (
    uid,
    payment_method_id,
    payment_request_id,
    payment_reference_id,
    payment_customer_id,
    payment_business_id,
    payment_type,
    payment_status,
    payment_reusability,
    payment_channel,
    payment_amount,
    payment_qr_code,
    payment_virtual_account_number,
    payment_url,
    payment_description,
    created_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetPaymentMethodByPaymentMethodID :one
SELECT * FROM payment_method WHERE payment_method_id = $1 LIMIT 1;

-- name: GetPaymentMethodByReferenceID :one
SELECT * FROM payment_method WHERE payment_reference_id = $1 LIMIT 1;

-- name: GetPaymentMethodCustomer :one
SELECT * FROM payment_method WHERE payment_method_id = $1 AND payment_customer_id = $2 LIMIT 1;

-- name: UpdatePaymentMethodCustomer :one
UPDATE payment_method
SET
    payment_status = COALESCE(sqlc.narg(payment_status), payment_status),
    payment_failure_code = COALESCE(sqlc.narg(payment_failure_code), payment_failure_code),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at),
    paid_at = COALESCE(sqlc.narg(paid_at), paid_at)
WHERE
    payment_method_id = sqlc.arg(payment_method_id)
AND
    payment_customer_id = sqlc.arg(payment_customer_id)
RETURNING *;