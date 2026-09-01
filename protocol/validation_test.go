package protocol

import "testing"

func TestValidateMessageRejectsOversizedText(t *testing.T) {
	long := make([]byte, MaxMessageLen+1)
	for i := range long {
		long[i] = 'a'
	}
	err := ValidateMessage(&Message{Type: "chat", Message: string(long)})
	if err == nil {
		t.Fatal("expected error for oversized message text")
	}
}

func TestValidateMessageAcceptsMaxSize(t *testing.T) {
	text := make([]byte, MaxMessageLen)
	for i := range text {
		text[i] = 'a'
	}
	if err := ValidateMessage(&Message{Type: "chat", Message: string(text)}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateField(t *testing.T) {
	if err := ValidateField("x", "ok", 10); err != nil {
		t.Fatalf("short value should pass: %v", err)
	}
	if err := ValidateField("x", "too long", 5); err == nil {
		t.Fatal("long value should fail")
	}
}
