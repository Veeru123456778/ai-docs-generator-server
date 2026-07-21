package service // Or package domain / errors

import "errors"

// Sentinel errors for standard service failures
var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrBlockNotFound    = errors.New("block not found")
	ErrEmptyBlocksBatch = errors.New("blocks payload cannot be empty")
)