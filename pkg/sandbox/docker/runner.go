package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/registry"
	fileUtils "github.com/anurag-327/neuron/internal/util/file"
	"github.com/anurag-327/neuron/pkg/logger"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// Runner is responsible for executing user-submitted code
// inside an already running (pooled) Docker container.
type Runner struct {
	client *conn.DockerClient
}

func NewRunner(client *conn.DockerClient) *Runner {
	return &Runner{client: client}
}

// RunResult represents the final outcome of a sandbox execution.
//
// ContainerDirty:
//   - true  → container must be destroyed & replaced
//   - false → container can be safely reused
type RunResult struct {
	Stdout         string
	Stderr         string
	ErrType        models.SandboxError
	ErrMsg         string
	ExitCode       int64
	ContainerDirty bool
}

// ------------------------------------------------------------
// Run
// ------------------------------------------------------------
//
// High-level execution flow:
//
// 1. Create a per-job directory on the HOST
// 2. Write user code + input into that directory
// 3. Execute code inside container using docker exec
// 4. Enforce TIME LIMIT using BusyBox `timeout` (inside container)
// 5. Use Go context timeout ONLY as a safety net
// 6. Classify result (TLE / MLE / RE / OK)
//
// IMPORTANT TIMEOUT DESIGN:
//
//		inner timeout  <  Go exec timeout
//	 Container is dirty only when go looses control over it and not when program times out correctly
//
// Example:
//
//	inner timeout = 2s   (authoritative TLE decision)
//	Go timeout    = 3s   (safety / cleanup)
//
// Exit-code policy (documented by design):
//
//	124 → timeout exited normally (TLE)
//	137 → SIGKILL (treated as TLE in this design)
//	139 → SIGSEGV (MLE)
func (d *Runner) Run(
	ctx context.Context,
	containerID,
	basePathString, code, input, language string,
) RunResult {

	log := func(format string, args ...any) {
		fmt.Printf("[RUN] "+format+"\n", args...)
	}

	result := RunResult{}
	result.ExitCode = 1

	log("START | container=%s language=%s", containerID, language)

	// 1 Create job directory on HOST
	projectRoot, _ := os.Getwd()
	basePath := filepath.Join(projectRoot, basePathString)

	log("Creating job directory: %s", basePath)

	if err := os.MkdirAll(basePath, 0777); err != nil {
		log("ERROR creating job dir: %v", err)
		appLogger := logger.GetGlobalLogger()
		appLogger.Error(ctx, time.Now(), "Failed to create job directory", map[string]interface{}{
			"containerID": containerID,
			"language":    language,
			"error":       err.Error(),
		})
		result.ErrType = models.ErrInternalError
		result.ErrMsg = "Failed to create job directory"
		return result
	}

	// FORCE 0777 to bypass permissions error
	if err := os.Chmod(basePath, 0777); err != nil {
		log("ERROR chmod job dir: %v", err)
		result.ErrType = models.ErrInternalError
		result.ErrMsg = "Failed to set permissions"
		return result
	}

	defer func() {
		log("Deleting job directory: %s", basePath)
		fileUtils.DeleteFolder(basePath)
	}()

	// 2 Load language configuration
	log("Loading language config: %s", language)

	langCfg, ok := registry.LanguageRegistry[language]
	if !ok {
		log("ERROR unsupported language")
		result.ErrType = models.ErrInternalError
		result.ErrMsg = "Unsupported language"
		return result
	}

	names := BuildFileNames(basePath, langCfg)

	// 3 Write user code and input
	log("Writing code file: %s", names.PathFull)

	if err := fileUtils.WriteContentToFile(names.PathFull, []byte(code), 0777); err != nil {
		log("ERROR writing code: %v", err)
		result.ErrType = models.ErrInternalError
		result.ErrMsg = "Failed to write code"
		return result
	}

	log("Writing input.txt")

	if err := fileUtils.WriteContentToFile(
		filepath.Join(basePath, "input.txt"),
		[]byte(input),
		0777,
	); err != nil {
		log("ERROR writing input: %v", err)
		result.ErrType = models.ErrInternalError
		result.ErrMsg = "Failed to write input"
		return result
	}

	containerJobPath := filepath.Join("/sandbox", filepath.Base(basePath))
	log("Container job path: %s", containerJobPath)

	// 4 Build execution command
	runCmd := langCfg.RunCmd(names)

	runTimeout := 3 * time.Second
	execTimeout := 4 * time.Second

	log("Timeouts | run=%s exec=%s", runTimeout, execTimeout)
	log("Run command: %s", runCmd)

	execCmd := []string{
		"sh", "-c",
		fmt.Sprintf(
			"cd %s && timeout -s KILL %ds sh -c '%s'",
			containerJobPath,
			int(runTimeout.Seconds()),
			runCmd,
		),
	}

	// 5 Create docker exec (NO timeout here)
	log("Creating docker exec")

	execResp, err := d.client.Client.ContainerExecCreate(
		context.Background(),
		containerID,
		container.ExecOptions{
			Cmd:          execCmd,
			AttachStdout: true,
			AttachStderr: true,
		},
	)
	if err != nil {
		log("ERROR exec create failed: %v", err)
		appLogger := logger.GetGlobalLogger()
		appLogger.Error(ctx, time.Now(), "Docker exec create failed", map[string]interface{}{
			"container_id": containerID,
			"language":     language,
			"error":        err.Error(),
		})
		result.ErrType = models.ErrSandboxError
		result.ErrMsg = "Exec create failed"
		result.ContainerDirty = true
		return result
	}

	log("Exec created: %s", execResp.ID)

	// 6 Attach to exec & wait with Go timeout
	execCtx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	deadline, _ := execCtx.Deadline()
	log("Attaching to exec | deadline=%v", deadline)

	attach, err := d.client.Client.ContainerExecAttach(
		execCtx,
		execResp.ID,
		container.ExecStartOptions{},
	)
	if err != nil {
		log("ERROR exec attach failed: %v", err)
		appLogger := logger.GetGlobalLogger()
		appLogger.Error(ctx, time.Now(), "Docker exec attach failed", map[string]interface{}{
			"container_id": containerID,
			"language":     language,
			"error":        err.Error(),
		})
		result.ErrType = models.ErrSandboxError
		result.ErrMsg = "Exec attach failed"
		result.ContainerDirty = true
		return result
	}
	defer attach.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	done := make(chan error, 1)

	go func() {
		log("Started stdout/stderr reader")
		_, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attach.Reader)
		done <- err
	}()

	// 7 Wait for completion OR Go-side timeout
	select {

	case <-execCtx.Done():
		log("GO TIMEOUT HIT | err=%v", execCtx.Err())
		appLogger := logger.GetGlobalLogger()
		appLogger.Error(ctx, time.Now(), "Execution timeout (Go context)", map[string]interface{}{
			"container_id": containerID,
			"language":     language,
			"timeout_sec":  execTimeout.Seconds(),
		})
		result.ErrType = models.ErrTLE
		result.ErrMsg = models.MsgTLE
		result.ContainerDirty = true
		return result

	case err := <-done:
		log("Exec finished | reader err=%v", err)
		if err != nil {
			result.ErrType = models.ErrSandboxError
			result.ErrMsg = "Output read failed"
			result.ContainerDirty = true
			return result
		}
	}

	// 8 Inspect exit code for classification
	inspect, _ := d.client.Client.ContainerExecInspect(
		context.Background(),
		execResp.ID,
	)

	log("Final inspect | pid=%d exit=%d",
		inspect.Pid, inspect.ExitCode)

	// Parse Error
	r := ProcessResult(language, int64(inspect.ExitCode), stdoutBuf.String(), stderrBuf.String(), containerJobPath)
	result.ErrType = r.ErrorType
	result.ErrMsg = r.ErrorMessage
	result.Stdout = r.Stdout
	result.Stderr = r.Stderr
	result.ExitCode = r.ExitCode

	if r.ExitCode == 139 || r.ExitCode == 124 || r.ExitCode == 137 {
		result.ContainerDirty = true
	}

	log("Execution completed successfully")
	return result
}

func (d *Runner) Health() error {
	_, err := d.client.Ping(context.Background())
	return err
}
