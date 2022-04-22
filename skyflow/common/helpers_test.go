package common

import "testing"

func TestAppendRequestId(t *testing.T) {
	var message = AppendRequestId("message", "1234")
	check(message, "message - requestId : 1234", t)
}

func check(got string, wanted string, t *testing.T) {
	if got != wanted {
		t.Errorf("got  %s, wanted %s", got, wanted)
	}
}
