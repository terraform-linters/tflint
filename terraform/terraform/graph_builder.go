package terraform

import (
	"log"

	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/logging"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
)

// GraphBuilder is an interface that can be implemented and used with
// Terraform to build the graph that Terraform walks.
type GraphBuilder interface {
	// Build builds the graph for the given module path. It is up to
	// the interface implementation whether this build should expand
	// the graph or not.
	Build(addrs.ModuleInstance) (*Graph, tfdiags.Diagnostics)
}

// BasicGraphBuilder is a GraphBuilder that builds a graph out of a
// series of transforms and (optionally) validates the graph is a valid
// structure.
type BasicGraphBuilder struct {
	Steps    []GraphTransformer
	Validate bool
	// Optional name to add to the graph debug log
	Name string
}

func (b *BasicGraphBuilder) Build(path addrs.ModuleInstance) (*Graph, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	g := &Graph{Path: path}

	var lastStepStr string
	for _, step := range b.Steps {
		if step == nil {
			continue
		}
		log.Printf("[TRACE] Executing graph transform %T", step)

		err := step.Transform(g)
		if thisStepStr := g.StringWithNodeTypes(); thisStepStr != lastStepStr {
			log.Printf("[TRACE] Completed graph transform %T with new graph:\n%s  ------", step, logging.Indent(thisStepStr))
			lastStepStr = thisStepStr
		} else {
			log.Printf("[TRACE] Completed graph transform %T (no changes)", step)
		}

		if err != nil {
			if nf, isNF := err.(tfdiags.NonFatalError); isNF {
				diags = diags.Append(nf.Diagnostics)
			} else {
				diags = diags.Append(err)
				return g, diags
			}
		}
	}

	// Validate the graph structure
	if b.Validate {
		if err := g.Validate(); err != nil {
			log.Printf("[ERROR] Graph validation failed. Graph:\n\n%s", g.String())
			diags = diags.Append(err)
			return nil, diags
		}
	}

	return g, diags
}
