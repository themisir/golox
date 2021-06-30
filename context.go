package main

import (
	"fmt"
	"os"
)

type LoxContext struct {
	hadError bool
}

func MakeContext() *LoxContext {
	return &LoxContext{
		hadError: false,
	}
}

func (c *LoxContext) error(line int, message string, a ...interface{}) {
	c.report(line, "", message, a...)
}

func (c *LoxContext) tokenError(token *Token, message string, a ...interface{}) {
	if token.tokenType == EOF {
		c.report(token.line, " at end", message, a...)
	} else {
		c.report(token.line, fmt.Sprintf(" at '%s'", token.lexme), message, a...)
	}
}

func (c *LoxContext) runtimeError(token *Token, message string, a ...interface{}) {
	c.tokenError(token, message, a...)
	panic(MakeRuntimeError(token, message, a...))
}

func (c *LoxContext) report(line int, where string, message string, a ...interface{}) {
	c.hadError = true
	message = fmt.Sprintf(message, a...)
	fmt.Fprintf(os.Stderr, "[line %v] Error%s: %s\n", line, where, message)
}

type RuntimeError struct {
	token   *Token
	message string
}

func MakeRuntimeError(token *Token, message string, a ...interface{}) RuntimeError {
	return RuntimeError{token: token, message: fmt.Sprintf(message, a...)}
}

func (e RuntimeError) Error() string {
	if e.token == nil {
		return "Runtime error: " + e.message
	}
	return fmt.Sprintf("Runtime error at line %v: %s", e.token.line, e.message)
}
