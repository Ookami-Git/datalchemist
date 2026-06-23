package generator

import "testing"

func TestString(t *testing.T) {
	for _, n := range []int{1, 16, 32} {
		value := String(n)
		if len(value) != n {
			t.Fatalf("String(%d) length = %d", n, len(value))
		}
	}
	if String(16) == String(16) {
		t.Fatal("two generated strings unexpectedly match")
	}
}
