package actions

import (
	"testing"

	"github.com/tinker-works/donsy/netomatic"
)

func TestReadLog_ShouldPollThePublicBoundedDaemonLog(t *testing.T) {
	client := &fakeClient{log: netomatic.ReadDaemonLogResponse{Lines: []string{"partial", "second"}, NextOffset: 25}}
	msg, ok := ReadLog(client, 6)().(LogLoadedMsg)
	if !ok || msg.Err != nil || len(msg.Lines) != 2 || msg.Next != 25 {
		t.Fatalf("unexpected log message: %#v", msg)
	}
	if client.logOffset != 6 || client.logLimit != netomatic.MaxDaemonLogLines {
		t.Fatalf("unexpected daemon-log request: offset=%d limit=%d", client.logOffset, client.logLimit)
	}
}

func TestReadLog_ShouldExposeMissingClientAsVisibleError(t *testing.T) {
	msg := ReadLog(nil, 0)().(LogLoadedMsg)
	if msg.Err == nil {
		t.Fatal("expected missing daemon client to be visible")
	}
}
