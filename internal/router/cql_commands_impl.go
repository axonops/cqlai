package router

import (
	"github.com/axonops/cqlai/internal/db"
)

// CqlCommandVisitorImpl is a concrete implementation of CqlCommandVisitor.
type CqlCommandVisitorImpl struct {
	session *db.Session
}

// NewCqlCommandVisitorImpl creates a new CqlCommandVisitorImpl.
func NewCqlCommandVisitorImpl(session *db.Session) *CqlCommandVisitorImpl {
	return &CqlCommandVisitorImpl{session: session}
}
