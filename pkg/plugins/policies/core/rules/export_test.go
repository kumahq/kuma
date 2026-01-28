package rules

// BuildRulesForTesting is like BuildRules but allows explicitly specifying useCliques for testing purposes.
func BuildRulesForTesting(list []PolicyItemWithMeta, withNegations bool, useCliques bool) (Rules, error) {
	return buildRulesInternal(list, withNegations, useCliques)
}
