-- name: CreatePaymentReusability :one
INSERT INTO payment_reusability (
    prname
) VALUES (
    $1
) RETURNING *;

-- name: GetPaymentReusabilityByName :one
SELECT * FROM payment_reusability WHERE prname = $1 LIMIT 1;