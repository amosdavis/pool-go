//go:build linux

package steps

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Format:      "pretty",
	Paths:       []string{"../features"},
	Strict:      false,
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			InitializePooliocScenario(ctx)
			InitializePoolScenario(ctx)
		},
		Options: &opts,
	}

	status := suite.Run()
	// status 0 = pass, 1 = fail, 2 = pending (non-strict)
	if status == 1 {
		t.Fatal("BDD scenarios failed")
	}
}
