package main

type FunctionType string

const (
	FUNCTION_NONE     FunctionType = "NONE"
	FUNCTION_FUNCTION FunctionType = "FUNCTION"
	FUNCTION_METHOD   FunctionType = "METHOD"
)

type Resolver struct {
	context         *LoxContext
	interpreter     *Interpreter
	scopes          *ResolverStack
	currentFunction FunctionType
}

func MakeResolver(context *LoxContext, interpreter *Interpreter) *Resolver {
	return &Resolver{
		context:         context,
		interpreter:     interpreter,
		scopes:          &ResolverStack{},
		currentFunction: FUNCTION_NONE,
	}
}

func (r *Resolver) visitBlockStmt(stmt *BlockStmt) Any {
	r.beginScope()
	r.resolve(stmt.statements)
	r.endScope()
	return nil
}

func (r *Resolver) resolve(statements []Stmt) {
	for _, statement := range statements {
		r.resolveStmt(statement)
	}
}

func (r *Resolver) resolveStmt(stmt Stmt) {
	stmt.accept(r)
}

func (r *Resolver) resolveExpr(expr Expr) {
	expr.accept(r)
}

func (r *Resolver) beginScope() {
	r.scopes.Push(make(map[string]bool))
}

func (r *Resolver) endScope() {
	r.scopes.Pop()
}

func (r *Resolver) visitVarStmt(stmt *VarStmt) Any {
	r.declare(stmt.name)
	if stmt.initializer != nil {
		r.resolveExpr(stmt.initializer)
	}
	r.define(stmt.name)
	return nil
}

func (r *Resolver) declare(name *Token) {
	if r.scopes.IsEmpty() {
		return
	}
	scope := r.scopes.Peek()
	if _, ok := scope[name.lexme]; ok {
		r.context.tokenError(name, "Already variable with this name in this scope.")
	}
	scope[name.lexme] = false
}

func (r *Resolver) define(name *Token) {
	if r.scopes.IsEmpty() {
		return
	}
	r.scopes.Peek()[name.lexme] = true
}

func (r *Resolver) visitVariableExpr(expr *VariableExpr) Any {
	if !r.scopes.IsEmpty() {
		if val, ok := r.scopes.Peek()[expr.name.lexme]; ok && !val {
			r.context.tokenError(expr.name, "Can't read local variable in its own initializer.")
		}
	}

	r.resolveLocal(expr, expr.name)
	return nil
}

func (r *Resolver) resolveLocal(expr Expr, name *Token) {
	for i := r.scopes.Size() - 1; i >= 0; i-- {
		if _, ok := r.scopes.Get(i)[name.lexme]; ok {
			r.interpreter.resolve(expr, r.scopes.Size()-1-i)
			return
		}
	}
}

func (r *Resolver) visitAssignExpr(expr *AssignExpr) Any {
	r.resolveExpr(expr.value)
	r.resolveLocal(expr, expr.name)
	return nil
}

func (r *Resolver) visitFunctionStmt(stmt *FunctionStmt) Any {
	r.declare(stmt.name)
	r.define(stmt.name)

	r.resolveFunction(stmt, FUNCTION_FUNCTION)
	return nil
}

func (r *Resolver) resolveFunction(function *FunctionStmt, functionType FunctionType) {
	enclosingFunction := r.currentFunction
	r.currentFunction = functionType

	r.beginScope()
	for _, param := range function.params {
		r.declare(param)
		r.define(param)
	}
	r.resolve(function.body)
	r.endScope()
	r.currentFunction = enclosingFunction
}

func (r *Resolver) visitExpressionStmt(stmt *ExpressionStmt) Any {
	r.resolveExpr(stmt.expression)
	return nil
}

func (r *Resolver) visitIfStmt(stmt *IfStmt) Any {
	r.resolveExpr(stmt.condition)
	r.resolveStmt(stmt.thenBranch)
	if stmt.elseBranch != nil {
		r.resolveStmt(stmt.elseBranch)
	}
	return nil
}

func (r *Resolver) visitReturnStmt(stmt *ReturnStmt) Any {
	if r.currentFunction == FUNCTION_NONE {
		r.context.tokenError(stmt.keyword, "Can't return from top-level code.")
	}

	if stmt.value != nil {
		r.resolveExpr(stmt.value)
	}
	return nil
}

func (r *Resolver) visitWhileStmt(stmt *WhileStmt) Any {
	r.resolveExpr(stmt.condition)
	r.resolveStmt(stmt.body)
	return nil
}

func (r *Resolver) visitBinaryExpr(expr *BinaryExpr) Any {
	r.resolveExpr(expr.left)
	r.resolveExpr(expr.right)
	return nil
}

func (r *Resolver) visitCallExpr(expr *CallExpr) Any {
	r.resolveExpr(expr.callee)
	for _, argument := range expr.arguments {
		r.resolveExpr(argument)
	}
	return nil
}

func (r *Resolver) visitGroupingExpr(expr *GroupingExpr) Any {
	r.resolveExpr(expr.expression)
	return nil
}

func (r *Resolver) visitLiteralExpr(expr *LiteralExpr) Any {
	return nil
}

func (r *Resolver) visitLogicalExpr(expr *LogicalExpr) Any {
	r.resolveExpr(expr.left)
	r.resolveExpr(expr.right)
	return nil
}

func (r *Resolver) visitUnaryExpr(expr *UnaryExpr) Any {
	r.resolveExpr(expr.right)
	return nil
}

func (r *Resolver) visitPrintStmt(stmt *PrintStmt) Any {
	r.resolveExpr(stmt.expression)
	return nil
}

func (r *Resolver) visitClassStmt(stmt *ClassStmt) Any {
	r.declare(stmt.name)

	for _, method := range stmt.methods {
		declaration := FUNCTION_METHOD
		r.resolveFunction(method, declaration)
	}

	r.define(stmt.name)
	return nil
}

func (r *Resolver) visitGetExpr(expr *GetExpr) Any {
	r.resolveExpr(expr.object)
	return nil
}

func (r *Resolver) visitSetExpr(expr *SetExpr) Any {
	r.resolveExpr(expr.object)
	r.resolveExpr(expr.value)
	return nil
}