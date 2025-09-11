package poll

import (
	"os"
	"testing"
)

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	if got := getEnvInt("TEST_INT", 0); got != 42 {
		t.Errorf("esperado 42, obteve %d", got)
	}
	os.Unsetenv("TEST_INT")
	if got := getEnvInt("TEST_INT", 99); got != 99 {
		t.Errorf("esperado fallback 99, obteve %d", got)
	}
}

func TestIsValidOption(t *testing.T) {
	options := []string{"A", "B", "C"}
	if !isValidOption(options, "A") {
		t.Error("A deveria ser válido")
	}
	if isValidOption(options, "X") {
		t.Error("X não deveria ser válido")
	}
}
