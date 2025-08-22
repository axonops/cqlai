package router

import (
	"github.com/antlr4-go/antlr/v4"
)

// CustomErrorListener collects syntax errors during parsing.
type CustomErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []string
}

// NewCustomErrorListener creates a new CustomErrorListener.
func NewCustomErrorListener() *CustomErrorListener {
	return new(CustomErrorListener)
}

// SyntaxError is called by ANTLR when a syntax error is found.
func (c *CustomErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	c.Errors = append(c.Errors, msg)
}
