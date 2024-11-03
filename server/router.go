package server

import "sync"

type Router struct {
	fn map[string]func(*Ctx)
	sync.RWMutex
}

func NewRouter() *Router {
	return &Router{
		fn: make(map[string]func(*Ctx)),
	}
}

func (r *Router) HandleFunc(cmd string, fn func(*Ctx)) {
	r.fn[cmd] = fn
}
