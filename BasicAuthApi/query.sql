-- name: CreateUser :one
INSERT INTO users (
    username, email, password_hash, security_question, security_answer_hash
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? LIMIT 1;

-- name: ListUsers :many
SELECT id, username, email, created_at FROM users;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = ? WHERE id = ?;