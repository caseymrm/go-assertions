package assertions

import (
	"encoding/json"
	"testing"
)

func TestSystemAssertions(t *testing.T) {
	a := GetAssertions()
	b, _ := json.MarshalIndent(a, "", "  ")
	_, ok := a["PreventUserIdleDisplaySleep"]
	if !ok {
		t.Errorf("%s\n", b)
	}
}

func TestPIDAssertions(t *testing.T) {
	a := GetPIDAssertions()
	b, _ := json.MarshalIndent(a, "", "  ")
	_, ok := a["PreventUserIdleDisplaySleep"]
	if !ok {
		t.Errorf("%s\n", b)
	}
}

func TestAssertionChanges(t *testing.T) {
	channel := make(chan AssertionChange)
	go func() {
		for change := range channel {
			t.Errorf("Change: %+v", change)
		}
	}()
	SubscribeAssertionChangesAndRun(channel)
}
