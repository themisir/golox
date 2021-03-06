package main

type Environment struct {
	context   *LoxContext
	values    map[string]Any
	enclosing *Environment
}

func MakeEnvironment(context *LoxContext, enclosing *Environment) *Environment {
	return &Environment{
		context:   context,
		values:    make(map[string]Any),
		enclosing: enclosing,
	}
}

func (e *Environment) extend() *Environment {
	return &Environment{
		context:   e.context,
		values:    make(map[string]Any),
		enclosing: e,
	}
}

func (e *Environment) define(name string, value Any) {
	e.values[name] = value
}

func (e *Environment) get(name *Token) Any {
	if val, ok := e.values[name.lexme]; ok {
		return val
	}

	if e.enclosing != nil {
		return e.enclosing.get(name)
	}

	e.context.runtimeError(name, "Undefined variable '%s'.", name.lexme)
	return nil // will not be executed
}

func (e *Environment) getAt(distance int, name string) Any {
	return e.ancestor(distance).values[name]
}

func (e *Environment) ancestor(distance int) *Environment {
	environment := e
	for i := 0; i < distance; i++ {
		environment = environment.enclosing
	}

	return environment
}

func (e *Environment) assign(name *Token, value Any) {
	if _, ok := e.values[name.lexme]; ok {
		e.values[name.lexme] = value
		return
	}

	if e.enclosing != nil {
		e.enclosing.assign(name, value)
		return
	}

	e.context.runtimeError(name, "Undefined variable '%s'.", name.lexme)
}

func (e *Environment) assignAt(distance int, name string, value Any) {
	e.ancestor(distance).values[name] = value
}
