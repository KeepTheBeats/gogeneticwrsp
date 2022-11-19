package experimenttools

import (
	"fmt"
	"github.com/KeepTheBeats/routing-algorithms/random"
	"testing"
)

func TestGenerateCloudsApps(t *testing.T) {
	var n int = 5
	for i := 0; i < n; i++ {
		numApps := random.RandomInt(3, 6)
		GenerateApps(numApps, fmt.Sprintf("%d", i))
	}
}
