package models

import "errors"

var (
	ErrTransaction      = errors.New("transaction error")
	ErrUserUpsert       = errors.New("failed to upsert student")
	ErrEligibility      = errors.New("eligibility check failed")
	ErrDocumentInsert   = errors.New("failed to insert document")
	ErrDocumentNotFound = errors.New("document not found")
	ErrStatusSet        = errors.New("failed to set status")
)
