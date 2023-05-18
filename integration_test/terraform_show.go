package test

import (
	"encoding/json"
	"log"

	tfjson "github.com/hashicorp/terraform-json"
)

/*
Parse the terraform show output, and marshal into a StateResource
*/
func parseShowJson(jsonStr string) map[string]*tfjson.StateResource {
	plan := &tfjson.State{}

	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		log.Fatal(err)
	}
	rootModule := plan.Values.RootModule
	return parseModulePlannedValues(rootModule)
}

// parseModulePlannedValues will recursively walk through the modules in the planned_values of the plan struct to
// construct a map that maps the full resource addresses to the planned resource.
func parseModulePlannedValues(module *tfjson.StateModule) map[string]*tfjson.StateResource {
	out := map[string]*tfjson.StateResource{}
	for _, resource := range module.Resources {
		// NOTE: the Address attribute of the module resource always returns the full address, even when the resource is
		// nested within sub modules.
		out[resource.Address] = resource
	}

	// NOTE: base case of recursion is when ChildModules is empty list.
	for _, child := range module.ChildModules {
		// Recurse in to the child module. We take a recursive approach here despite limitations of the recursion stack
		// in golang due to the fact that it is rare to have heavily deep module calls in Terraform. So we optimize for
		// code readability as opposed to performance.
		childMap := parseModulePlannedValues(child)
		for k, v := range childMap {
			out[k] = v
		}
	}
	return out
}
