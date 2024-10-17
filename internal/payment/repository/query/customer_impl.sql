-- name: CreateCustomer :one
INSERT INTO customer (
    uid,
    customer_app_id,
    payment_customer_id,
    customer_name,
    created_at,
    email,
    phone_number
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetCustomerByPaymentCustomerID :one
SELECT * FROM customer WHERE payment_customer_id = $1 LIMIT 1;

-- name: GetCustomerByCustomerAppID :one
SELECT * FROM customer WHERE customer_app_id = $1 LIMIT 1;