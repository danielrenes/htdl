package transform

import (
	"context"

	"github.com/danielrenes/htdl/internal/html"
)

type TransformerContext struct {
	ctx context.Context
}

func NewTransformerContext() *TransformerContext {
	return &TransformerContext{ctx: context.Background()}
}

func (t *TransformerContext) GetValue(key any) any {
	return t.ctx.Value(key)
}

func (t *TransformerContext) SetValue(key, value any) {
	t.ctx = context.WithValue(t.ctx, key, value)
}

type Transformer interface {
	Transform(node *html.Node, ctx *TransformerContext) error
}

type TransformerFunc func(node *html.Node, ctx *TransformerContext) error

func (fn TransformerFunc) Transform(node *html.Node, ctx *TransformerContext) error {
	return fn(node, ctx)
}
