package tfexec

import "log"

// TerraformPlanJSON represents the Terraform plan in JSON format
type TerraformPlanJSON struct {
	FormatVersion   string                  `json:"format_version"`
	Applyable       bool                    `json:"applyable"`
	Complete        bool                    `json:"complete"`
	Errored         bool                    `json:"errored"`
	ResourceChanges []ResourceChange        `json:"resource_changes"`
	OutputChanges   map[string]OutputChange `json:"output_changes"`
}

// ResourceChange represents a change to a resource in the plan
type ResourceChange struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address,omitempty"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Index         *int   `json:"index,omitempty"`
	Deposed       string `json:"deposed,omitempty"`
	Change        Change `json:"change"`
	ActionReason  string `json:"action_reason,omitempty"`
}

// OutputChange represents a change to an output value
type OutputChange struct {
	Change Change `json:"change"`
}

// Change represents the change details (before, after, actions)
type Change struct {
	Actions []string    `json:"actions"`
	Before  interface{} `json:"before"`
	After   interface{} `json:"after"`
}

// HasChanges returns true if there are any resource changes in the plan
func (p *TerraformPlanJSON) HasChanges() bool {
	hasChanges := false
	for _, rc := range p.ResourceChanges {
		// "no-op" means no changes - all other actions indicate changes
		if len(rc.Change.Actions) != 1 || rc.Change.Actions[0] != "no-op" {
			// Check if changes are only to tags
			onlyTagChanges := false
			if len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "update" {
				before, beforeOk := rc.Change.Before.(map[string]interface{})
				after, afterOk := rc.Change.After.(map[string]interface{})

				if beforeOk && afterOk {
					// Assume changes are only to tags unless we find otherwise
					onlyTagChanges = true

					// Compare all keys except "tags"
					for k, v := range before {
						if k != "tags" {
							if afterVal, exists := after[k]; !exists || afterVal != v {
								onlyTagChanges = false
								break
							}
						}
					}

					for k := range after {
						if k != "tags" && before[k] == nil {
							onlyTagChanges = false
							break
						}
					}
				}
			}

			if !onlyTagChanges {
				log.Printf("Change detected in resource: %s, actions: %v", rc.Address, rc.Change.Actions)
				hasChanges = true
			} else {
				log.Printf("Ignoring tag-only changes in resource: %s", rc.Address)
			}
		}
	}
	return hasChanges
}

// HasOnlyOutputChanges returns true if there are only output changes and no resource changes
func (p *TerraformPlanJSON) HasOnlyOutputChanges() bool {
	hasOutputChanges := len(p.OutputChanges) > 0

	// Check if there are any resource changes
	for _, rc := range p.ResourceChanges {
		// Any action other than "no-op" is a resource change
		if len(rc.Change.Actions) != 1 || rc.Change.Actions[0] != "no-op" {
			return false
		}
	}

	return hasOutputChanges
}

func (p *TerraformPlanJSON) LogResourceChanges() {
	for _, rc := range p.ResourceChanges {
		// Skip resources with "no-op" actions
		if len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "no-op" {
			continue
		}

		log.Printf("Resource Change: Address=%s, Type=%s, Name=%s, Mode=%s, Actions=%v",
			rc.Address, rc.Type, rc.Name, rc.Mode, rc.Change.Actions)
		if rc.Change.Before != nil {
			log.Printf("  Before: %v", rc.Change.Before)
		}
		if rc.Change.After != nil {
			log.Printf("  After: %v", rc.Change.After)
		}
	}
}
