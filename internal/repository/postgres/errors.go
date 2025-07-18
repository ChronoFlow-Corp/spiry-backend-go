package postgres

type errorCode = string

const (
	uniqueViolation errorCode = "unique_violation"
)