package terraform

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
)

// NodeCountBoundary fixes up any transitions between "each modes" in objects
// saved in state, such as switching from NoEach to EachInt.
type NodeCountBoundary struct {
	Config *configs.Config
}

var _ GraphNodeExecutable = (*NodeCountBoundary)(nil)

func (n *NodeCountBoundary) Name() string {
	return "meta.count-boundary (EachMode fixup)"
}

// GraphNodeExecutable
func (n *NodeCountBoundary) Execute(ctx EvalContext, op walkOperation) (diags tfdiags.Diagnostics) {
	// We'll temporarily lock the state to grab the modules, then work on each
	// one separately while taking a lock again for each separate resource.
	// This means that if another caller concurrently adds a module here while
	// we're working then we won't update it, but that's no worse than the
	// concurrent writer blocking for our entire fixup process and _then_
	// adding a new module, and in practice the graph node associated with
	// this eval depends on everything else in the graph anyway, so there
	// should not be concurrent writers.
	state := ctx.State().Lock()
	moduleAddrs := make([]addrs.ModuleInstance, 0, len(state.Modules))
	for _, m := range state.Modules {
		moduleAddrs = append(moduleAddrs, m.Addr)
	}
	ctx.State().Unlock()

	for _, addr := range moduleAddrs {
		cfg := n.Config.DescendentForInstance(addr)
		if cfg == nil {
			log.Printf("[WARN] Not fixing up EachModes for %s because it has no config", addr)
			continue
		}
		if err := n.fixModule(ctx, addr); err != nil {
			diags = diags.Append(err)
			return diags
		}
	}
	return diags
}

func (n *NodeCountBoundary) fixModule(ctx EvalContext, moduleAddr addrs.ModuleInstance) error {
	ms := ctx.State().Module(moduleAddr)
	cfg := n.Config.DescendentForInstance(moduleAddr)
	if ms == nil {
		// Theoretically possible for a concurrent writer to delete a module
		// while we're running, but in practice the graph node that called us
		// depends on everything else in the graph and so there can never
		// be a concurrent writer.
		return fmt.Errorf("[WARN] no state found for %s while trying to fix up EachModes", moduleAddr)
	}
	if cfg == nil {
		return fmt.Errorf("[WARN] no config found for %s while trying to fix up EachModes", moduleAddr)
	}

	for _, r := range ms.Resources {
		rCfg := cfg.Module.ResourceByAddr(r.Addr.Resource)
		if rCfg == nil {
			log.Printf("[WARN] Not fixing up EachModes for %s because it has no config", r.Addr)
			continue
		}
		hasCount := rCfg.Count != nil
		fixResourceCountSetTransition(ctx, r.Addr.Config(), hasCount)
	}

	return nil
}
