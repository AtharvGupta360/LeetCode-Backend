package judge

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// languageConfig holds all Docker + execution settings for one language.
type languageConfig struct {
	Image    string // Docker image to pull/use
	FileName string // file to write user code into inside the container
	Command  []string // command to compile+run (or just run)
}

// supportedLanguages maps language name → its execution config.
var supportedLanguages = map[string]languageConfig{
	"python": {
		Image:    "python:3.12-slim",
		FileName: "solution.py",
		Command:  []string{"python3", "/code/solution.py"},
	},
	"go": {
		Image:    "golang:1.22-alpine",
		FileName: "solution.go",
		Command:  []string{"sh", "-c", "cd /code && go run solution.go"},
	},
	"cpp": {
		Image:    "gcc:13",
		FileName: "solution.cpp",
		Command:  []string{"sh", "-c", "g++ -O2 -o /code/solution /code/solution.cpp && /code/solution"},
	},
	"java": {
		Image:    "openjdk:21-slim",
		FileName: "Solution.java",
		Command:  []string{"sh", "-c", "javac /code/Solution.java && java -cp /code Solution"},
	},
	"javascript": {
		Image:    "node:20-alpine",
		FileName: "solution.js",
		Command:  []string{"node", "/code/solution.js"},
	},
}

// RunResult holds the outcome of a single test case execution.
type RunResult struct {
	Stdout    string
	Stderr    string
	RuntimeMs int64
	TimedOut  bool
	OOMKilled bool
}

// Runner executes user code inside isolated Docker containers.
type Runner struct {
	client        *client.Client
	timeLimitSec  int
	memLimitBytes int64
}

// NewRunner creates a Runner connected to the local Docker daemon.
func NewRunner(timeLimitSec int, memLimitMB int64) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	return &Runner{
		client:        cli,
		timeLimitSec:  timeLimitSec,
		memLimitBytes: memLimitMB * 1024 * 1024,
	}, nil
}

// Run executes user code for one test case input and returns the result.
func (r *Runner) Run(ctx context.Context, language, code, stdin string) (*RunResult, error) {
	cfg, ok := supportedLanguages[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Create a context with time limit so the container is killed on timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(r.timeLimitSec)*time.Second)
	defer cancel()

	// 1. Create the container (don't start yet)
	resp, err := r.client.ContainerCreate(
		timeoutCtx,
		&container.Config{
			Image:        cfg.Image,
			Cmd:          cfg.Command,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			StdinOnce:    true,
			NetworkDisabled: true, // NO internet access
		},
		&container.HostConfig{
			Resources: container.Resources{
				Memory:     r.memLimitBytes,
				MemorySwap: r.memLimitBytes, // prevent swap usage
				CPUQuota:   50000,           // 50% of one CPU
				CPUPeriod:  100000,
				PidsLimit:  func(i int64) *int64 { return &i }(64), // max 64 processes
			},
			ReadonlyRootfs: true, // container filesystem is read-only
			Tmpfs: map[string]string{
				"/code": "size=10m", // writable temp dir for code only
				"/tmp":  "size=5m",
			},
		},
		nil, nil, "",
	)
	if err != nil {
		return nil, fmt.Errorf("container create failed: %w", err)
	}
	containerID := resp.ID

	// Always clean up the container when done
	defer r.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}) //nolint:errcheck

	// 2. Copy user code into /code/<filename> inside the container
	if err := r.copyCodeToContainer(timeoutCtx, containerID, cfg.FileName, code); err != nil {
		return nil, fmt.Errorf("copy code failed: %w", err)
	}

	// 3. Attach to the container to stream stdin/stdout/stderr
	hijack, err := r.client.ContainerAttach(timeoutCtx, containerID, container.AttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return nil, fmt.Errorf("container attach failed: %w", err)
	}
	defer hijack.Close()

	// 4. Start the container
	start := time.Now()
	if err := r.client.ContainerStart(timeoutCtx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("container start failed: %w", err)
	}

	// 5. Write stdin (test case input) and close it
	if _, err := io.WriteString(hijack.Conn, stdin); err != nil {
		return nil, fmt.Errorf("stdin write failed: %w", err)
	}
	hijack.CloseWrite() //nolint:errcheck

	// 6. Read stdout + stderr from the multiplexed stream
	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := stdcopy(&stdoutBuf, &stderrBuf, hijack.Reader); err != nil && err != io.EOF {
		// Timeout context cancellation shows up as an error here — that's expected
		if timeoutCtx.Err() != context.DeadlineExceeded {
			return nil, fmt.Errorf("output read failed: %w", err)
		}
	}

	// 7. Wait for the container to finish
	statusCh, errCh := r.client.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)
	var exitCode int64
	var oomKilled bool
	select {
	case status := <-statusCh:
		exitCode = status.StatusCode
		// Check OOM kill
		inspect, err := r.client.ContainerInspect(ctx, containerID)
		if err == nil && inspect.State != nil {
			oomKilled = inspect.State.OOMKilled
		}
	case err := <-errCh:
		if err != nil && timeoutCtx.Err() != context.DeadlineExceeded {
			return nil, fmt.Errorf("container wait failed: %w", err)
		}
	case <-timeoutCtx.Done():
		// Time limit exceeded — force-stop the container
		r.client.ContainerKill(ctx, containerID, "SIGKILL") //nolint:errcheck
		return &RunResult{
			Stdout:    stdoutBuf.String(),
			Stderr:    stderrBuf.String(),
			RuntimeMs: int64(time.Since(start).Milliseconds()),
			TimedOut:  true,
		}, nil
	}

	_ = exitCode // may use for compile error detection in future

	return &RunResult{
		Stdout:    strings.TrimSpace(stdoutBuf.String()),
		Stderr:    strings.TrimSpace(stderrBuf.String()),
		RuntimeMs: int64(time.Since(start).Milliseconds()),
		OOMKilled: oomKilled,
	}, nil
}

// copyCodeToContainer writes user code as a tar archive into /code/ inside the container.
// Docker's CopyToContainer API requires a tar stream, not a raw file.
func (r *Runner) copyCodeToContainer(ctx context.Context, containerID, filename, code string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte(code)
	if err := tw.WriteHeader(&tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(content)),
	}); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	tw.Close()

	return r.client.CopyToContainer(ctx, containerID, "/code", &buf, container.CopyToContainerOptions{})
}

// stdcopy demultiplexes Docker's multiplexed stdout/stderr stream.
// Docker prefixes each chunk with an 8-byte header: [stream_type, 0, 0, 0, size(4 bytes)].
func stdcopy(dst, dstErr io.Writer, src io.Reader) (int64, error) {
	hdr := make([]byte, 8)
	var written int64
	for {
		if _, err := io.ReadFull(src, hdr); err != nil {
			return written, err
		}
		streamType := hdr[0]
		frameSize := int64(hdr[4])<<24 | int64(hdr[5])<<16 | int64(hdr[6])<<8 | int64(hdr[7])

		var w io.Writer
		if streamType == 2 {
			w = dstErr
		} else {
			w = dst
		}
		n, err := io.CopyN(w, src, frameSize)
		written += n
		if err != nil {
			return written, err
		}
	}
}
