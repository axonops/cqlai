package cluster

import "errors"

// Common errors returned by the metadata manager
var (
	// ErrKeyspaceNotFound indicates the keyspace does not exist
	ErrKeyspaceNotFound = errors.New("keyspace not found")

	// ErrTableNotFound indicates the table does not exist
	ErrTableNotFound = errors.New("table not found")

	// ErrTypeNotFound indicates the UDT does not exist
	ErrTypeNotFound = errors.New("user type not found")

	// ErrInvalidMetadata indicates metadata is malformed or invalid
	ErrInvalidMetadata = errors.New("invalid metadata")

	// ErrUnsupportedFeature indicates feature not available in this Cassandra version
	ErrUnsupportedFeature = errors.New("unsupported feature")

	// ErrVectorParseFailed indicates vector type dimension/type parse failed
	ErrVectorParseFailed = errors.New("vector type parse failed")

	// ErrInvalidVectorDimension indicates vector dimension is out of valid range
	ErrInvalidVectorDimension = errors.New("invalid vector dimension")
)
