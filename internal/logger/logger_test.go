package logger

import (
	"context"
	"testing"
)

func TestTraceRoundTrip(t *testing.T) {
	ctx := WithTrace(context.Background(), "trace-123")
	if got := TraceFrom(ctx); got != "trace-123" {
		t.Fatalf("want trace-123 got %q", got)
	}
}

func TestTraceFromMissing(t *testing.T) {
	if got := TraceFrom(context.Background()); got != "" {
		t.Fatalf("want empty got %q", got)
	}
}

func TestFromContextFallsBackToGlobal(t *testing.T) {
	l := FromContext(context.Background())
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestConfigureLevel(t *testing.T) {
	Configure("debug")
	if Logger == nil {
		t.Fatal("expected logger after configure")
	}
	Configure("invalid")
	if Logger == nil {
		t.Fatal("expected logger after invalid level")
	}
}
