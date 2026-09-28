package logs

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestInit(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("LOG_LEVEL", "info")
	initd = false

	Init()
	Init() // idempotent

	name := time.Now().UTC().Format(time.DateOnly) + ".log"
	log.Info().Msg("to file")
	if b, err := os.ReadFile(name); err != nil || !strings.Contains(string(b), "to file") {
		t.Errorf("log file %s = %q, %v", name, b, err)
	}
}

// capture returns everything written to stdout by a logger built while fn runs.
func capture(t *testing.T, level string, fn func(zerolog.Logger)) string {
	t.Helper()
	t.Setenv("LOG_LEVEL", level)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	logger := MakeZerolog()
	os.Stdout = stdout

	fn(logger)
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestMakeZerologLevels(t *testing.T) {
	out := capture(t, "", func(l zerolog.Logger) {
		l.Trace().Msg("t")
		l.Debug().Msg("d")
		l.Info().Msg("i")
		l.Warn().Msg("w")
		l.Error().Msg("e")
		l.Log().Msg("n")
		func() {
			defer func() { _ = recover() }()
			l.Panic().Msg("p")
		}()
	})

	for _, want := range []string{
		Cyan + "TRA", Purple + "DEB", Green + "INF", Yellow + "WAR", Red + "ERR", " 🦁", "\033[41m\u001B[0;37mPAN",
		"logs_test.go", // trace level adds the caller
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestMakeZerologParsedLevel(t *testing.T) {
	out := capture(t, "warn", func(l zerolog.Logger) {
		l.Info().Msg("hidden")
		l.Warn().Msg("shown")
	})

	if strings.Contains(out, "hidden") || !strings.Contains(out, "shown") {
		t.Errorf("unexpected output for warn level:\n%s", out)
	}
	if strings.Contains(out, "logs_test.go") {
		t.Errorf("caller should only be added at trace level:\n%s", out)
	}
}
