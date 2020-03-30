package tags

import (
	"fmt"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
)

func MatchingRegex(tags v1alpha1.SingleValueTagSet) (re string) {
	for _, key := range tags.Keys() {
		keyIsEqual := fmt.Sprintf(`&%s=`, key)
		value := fmt.Sprintf(`[^&]*%s[,&]`, tags[key])
		expr := keyIsEqual + value + `.*`
		re += expr
	}
	return
}
