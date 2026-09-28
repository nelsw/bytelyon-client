package play

import (
	"errors"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// job records how it was run; its name picks the queue, and cmd is the command it runs.
type job struct {
	name, cmd string
	args      []string
	valid     bool
	validated atomic.Int32
	done      chan struct{}
	out       []byte
	err       error
}

func newJob(name, cmd string, valid bool, args ...string) *job {
	return &job{name: name, cmd: cmd, args: args, valid: valid, done: make(chan struct{})}
}

func (j *job) Name() string       { return j.name }
func (j *job) Args() []string     { return j.args }
func (j *job) Success(out []byte) { j.out = out; close(j.done) }
func (j *job) Failure(err error)  { j.err = err; close(j.done) }
func (j *job) Validate() bool     { j.validated.Add(1); return j.valid }
func (j *job) wait(t *testing.T) *job {
	t.Helper()
	select {
	case <-j.done:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s did not run", j.name)
	}
	return j
}

// commands runs every job's cmd in place of its name, so the queue routing can be tested with real commands.
func commands(t *testing.T) {
	t.Helper()
	orig := command
	command = func(p Playable) *exec.Cmd { return exec.Command(p.(*job).cmd, p.Args()...) }
	t.Cleanup(func() { command = orig })
}

func TestPlay(t *testing.T) {
	commands(t)
	t.Cleanup(func() { // reopen the queues this test closes
		closed = false
		start()
	})

	// Go and It route to each queue by name, validating each job exactly once
	for _, name := range []string{"./scripts/sync_x", "./scripts/async_x", "./scripts/x"} {
		j := newJob(name, "echo", true, "hi")
		Go(j)
		if j.wait(t).err != nil || string(j.out) != "hi\n" || j.validated.Load() != 1 {
			t.Errorf("Go(%s): out = %q, err = %v, validated %d times", name, j.out, j.err, j.validated.Load())
		}

		j = newJob(name, "false", true)
		It(j)
		var exit *exec.ExitError
		if !errors.As(j.wait(t).err, &exit) || j.validated.Load() != 1 {
			t.Errorf("It(%s): err = %v, validated %d times", name, j.err, j.validated.Load())
		}
	}

	// invalid jobs are dropped
	invalid := newJob("./scripts/sync_x", "echo", false)
	Go(invalid)
	It(invalid)
	if invalid.validated.Load() != 2 {
		t.Errorf("invalid job validated %d times, want 2", invalid.validated.Load())
	}

	// Close waits for accepted jobs, then drops new ones; calling it again is a no-op
	var slow []*job
	for range 10 {
		j := newJob("./scripts/sync_x", "sleep", true, "0.05")
		slow = append(slow, j)
		Go(j)
	}
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(Close)
	}
	wg.Wait()
	for _, j := range slow {
		select {
		case <-j.done:
		default:
			t.Fatal("Close returned before an accepted job finished")
		}
	}

	late := newJob("./scripts/sync_x", "echo", true)
	Go(late)
	It(late)
	if late.validated.Load() != 0 {
		t.Error("jobs should be dropped, unvalidated, after Close")
	}
}
