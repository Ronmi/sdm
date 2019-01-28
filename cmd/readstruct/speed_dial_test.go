package readstruct

import "testing"

func TestOKFindFirstIn(t *testing.T) {
	i, err := FindFirstIn("testdata/", ByName("A"))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if i.Name != "A" {
		t.Fatalf("Expected find strcut A, got %s", i.Name)
	}
}

func TestFailFindFirstIn(t *testing.T) {
	_, err := FindFirstIn("testdata/", ByName("X"))
	if err == nil {
		t.Fatal("Expected to get not found error, got nothing")
	}

	if err.Error() != "Cannot find matching structure." {
		t.Fatalf("Expected to get not found error, got %s", err)
	}
}
