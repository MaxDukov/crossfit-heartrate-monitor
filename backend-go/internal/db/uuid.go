package db

import "github.com/google/uuid"

// NewUUID генерирует UUID-строку (36 символов), как Python uuid4.
func NewUUID() string { return uuid.New().String() }
