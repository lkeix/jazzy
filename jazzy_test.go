package jazzy

import "testing"

func TestNew(t *testing.T) {
	jazzy := New()
	if jazzy == nil {
		t.Errorf("jazzy.New return nil")
	}
}

// TODO http method test
