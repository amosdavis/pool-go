//go:build linux

package steps

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
	Paths:  []string{"features"},
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			InitializePooliocScenario(ctx)
			InitializePoolScenario(ctx)
		},
		Options: &opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero exit status from godog")
	}
}
