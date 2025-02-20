package transform

import (
	"fmt"

	"github.com/danielrenes/htdl/internal/html"
)

type Pipeline struct {
	ctx          *TransformerContext
	transformers []Transformer
}

func NewPipeline(transformers ...Transformer) *Pipeline {
	return &Pipeline{ctx: NewTransformerContext(), transformers: transformers}
}

func (p *Pipeline) Run(node *html.Node) error {
	for _, transformer := range p.transformers {
		if err := transformer.Transform(node, p.ctx); err != nil {
			switch transformer := transformer.(type) {
			case *namedTransformer:
				return fmt.Errorf("%s: %w", transformer.name, err)
			default:
				return err
			}
		}
	}
	return nil
}

type namedTransformer struct {
	name        string
	transformer Transformer
}

func (t *namedTransformer) Transform(node *html.Node, ctx *TransformerContext) error {
	return t.transformer.Transform(node, ctx)
}

func Named(name string, transformer Transformer) Transformer {
	return &namedTransformer{name: name, transformer: transformer}
}
