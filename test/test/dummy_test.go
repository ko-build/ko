package test

import "testing"

func TestDummy(t *testing.T) {
	t.Run("dummy test", func(t *testing.T) {
		t.Logf("this is a dummy test")
	})
}
