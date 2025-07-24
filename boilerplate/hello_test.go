package boilerplate

import "testing"

func TestHello(t *testing.T) {
	t.Parallel()

	if want, got := "Hello", Hello(); want != got {
		t.Errorf("want = %s, god = %s", want, got)
	}
}
