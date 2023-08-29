package rpc

import (
	"context"
	"framework/network"

	"golang.org/x/exp/maps"
)

type Context interface {
	context.Context
	SetNetConn(conn network.Connection) Context
	GetNetConn() network.Connection
}

type ctxAttrConn struct{}

func createContext(parent context.Context) Context {
	ctx := &ContextImpl{
		Context: parent,
	}
	return ctx
}

type ContextImpl struct {
	context.Context
	attrib map[any]any
}

// call clone to avoid race condition
func (ctx *ContextImpl) clone() *ContextImpl {
	cc := &ContextImpl{
		Context: ctx.Context,
	}
	cc.attrib = maps.Clone(ctx.attrib)
	return cc
}

func (ctx *ContextImpl) SetAttrib(k, v any) *ContextImpl {
	newCtx := ctx.clone()
	if newCtx.attrib == nil {
		newCtx.attrib = make(map[any]any)
	}
	newCtx.attrib[k] = v
	return newCtx
}

func (ctx *ContextImpl) GetAttrib(k any) (any, bool) {
	if ctx.attrib == nil {
		return nil, false
	}
	v, ok := ctx.attrib[k]
	return v, ok
}

func (ctx *ContextImpl) SetNetConn(conn network.Connection) Context {
	return ctx.SetAttrib(ctxAttrConn{}, conn)
}

func (ctx *ContextImpl) GetNetConn() network.Connection {
	conn, ok := ctx.GetAttrib(ctxAttrConn{})
	if !ok {
		return nil
	}
	return conn.(network.Connection)
}
