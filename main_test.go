package main

import (
	"testing"
)

var config = &Config{
	Scripts: []*Script{
		{"success", "exit 0", 1},
		{"failure", "exit 255", 1},
		{"timeout", "sleep 5", 2},
	},
}

func TestRunScripts(t *testing.T) {
	measurements := runScripts(config.Scripts)

	expectedResults := map[string]struct {
		finished    int
		exitStatus  int
		minDuration float64
	}{
		"success": {1, 0, 0},
		"failure": {1, 255, 0},
		"timeout": {0, -1, 2},
	}

	for _, measurement := range measurements {
		expectedResult := expectedResults[measurement.Script.Name]

		if measurement.Finished != expectedResult.finished {
			t.Errorf("Expected result not found: %s", measurement.Script.Name)
		}

		if measurement.ExitStatus != expectedResult.exitStatus {
			t.Errorf("Expected exit status: %d, %d (Result)", measurement.ExitStatus, expectedResult.exitStatus)
		}

		if measurement.Duration < expectedResult.minDuration {
			t.Errorf("Expected duration %f < %f: %s", measurement.Duration, expectedResult.minDuration, measurement.Script.Name)
		}
	}
}

func TestScriptFilter(t *testing.T) {
	t.Run("RequiredParameters", func(t *testing.T) {
		_, err := scriptFilter(config.Scripts, "", "")

		if err.Error() != "`name` or `pattern` required" {
			t.Errorf("Expected failure when supplying no parameters")
		}
	})

	t.Run("NameMatch", func(t *testing.T) {
		scripts, err := scriptFilter(config.Scripts, "success", "")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 1 || scripts[0] != config.Scripts[0] {
			t.Errorf("Expected script not found")
		}
	})

	t.Run("PatternMatch", func(t *testing.T) {
		scripts, err := scriptFilter(config.Scripts, "", "fail.*")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 1 || scripts[0] != config.Scripts[1] {
			t.Errorf("Expected script not found")
		}
	})

	t.Run("AllMatch", func(t *testing.T) {
		scripts, err := scriptFilter(config.Scripts, "success", ".*")

		if err != nil {
			t.Errorf("Unexpected: %s", err.Error())
		}

		if len(scripts) != 3 {
			t.Fatalf("Expected 3 scripts, received %d", len(scripts))
		}

		for i, script := range config.Scripts {
			if scripts[i] != script {
				t.Fatalf("Expected script not found")
			}
		}
	})
}
