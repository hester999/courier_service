package pkg

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func newTestLogger(level Level) (*LoggerImpl, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := &LoggerImpl{
		level:  level,
		writer: log.New(buf, "", 0),
	}
	return logger, buf
}

func TestLogger_InfoPrintedWhenLevelAllows(t *testing.T) {
	logger, buf := newTestLogger(INFO)

	logger.Info("hello %s", "world")

	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Fatalf("expected INFO level in output, got %q", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected message in output, got %q", out)
	}
}

func TestLogger_InfoNotPrintedWhenLevelHigher(t *testing.T) {
	logger, buf := newTestLogger(WARN)

	logger.Info("should not be printed")

	if buf.Len() != 0 {
		t.Fatalf("expected no output, got %q", buf.String())
	}
}

func TestLogger_WarnPrinted(t *testing.T) {
	logger, buf := newTestLogger(INFO)

	logger.Warn("warn message")

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Fatalf("expected WARN level in output, got %q", out)
	}
	if !strings.Contains(out, "warn message") {
		t.Fatalf("expected message in output, got %q", out)
	}
}

func TestLogger_ErrorPrinted(t *testing.T) {
	logger, buf := newTestLogger(INFO)

	logger.Error("error occurred")

	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("expected ERROR level in output, got %q", out)
	}
}

func TestLogger_MonitoringPrintedOnlyWhenEnabled(t *testing.T) {
	logger, buf := newTestLogger(MONITORING)

	logger.Monitoring("metrics ok")

	out := buf.String()
	if !strings.Contains(out, "MONITORING") {
		t.Fatalf("expected MONITORING level, got %q", out)
	}
}

func TestLogger_MonitoringAlwaysPrinted(t *testing.T) {
	logger, buf := newTestLogger(ERROR)

	logger.Monitoring("metrics")

	out := buf.String()
	if !strings.Contains(out, "MONITORING") {
		t.Fatalf("expected MONITORING output, got %q", out)
	}
}
