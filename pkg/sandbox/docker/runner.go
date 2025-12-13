package docker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

type Runner struct {
	client *conn.DockerClient
}

func NewRunner(client *conn.DockerClient) *Runner {
	return &Runner{client: client}
}

// -----------------------------------------------------------------------------
//
//	Run: Executes code in Docker Sandbox
//
// -----------------------------------------------------------------------------
//
//	Returns:
//	  stdout, stderr, errType, errMessage
//
// -----------------------------------------------------------------------------
func (d *Runner) Run(
	ctx context.Context,
	containerID,
	basePathString, code, input, language string,
) (string, string, sandbox.SandboxError, string) {

	// 1️ Create job directory on HOST
	projectRoot, _ := os.Getwd()
	basePath := filepath.Join(projectRoot, basePathString)

	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Println("[RUN] failed to create job dir:", err)
		return "", "", sandbox.ErrInternalError, "Failed to create job directory"
	}
	defer util.DeleteFolder(basePath)

	log.Println("[RUN] job dir:", basePath)

	// 2️ Load language config
	languageConfig, err := GetLanguageConfig(language)
	if err != nil {
		log.Println("[RUN] unsupported language:", language)
		return "", "", sandbox.ErrInternalError, "Unsupported language"
	}

	names := BuildFileNames(basePath, languageConfig)

	// 3 Write code + input files
	if err := util.WriteContentToFile(names.PathFull, []byte(code), 0644); err != nil {
		log.Println("[RUN] failed writing code:", err)
		return "", "", sandbox.ErrInternalError, "Failed to write code file"
	}

	inputPath := filepath.Join(basePath, "input.txt")
	if err := util.WriteContentToFile(inputPath, []byte(input), 0644); err != nil {
		log.Println("[RUN] failed writing input:", err)
		return "", "", sandbox.ErrInternalError, "Failed to write input file"
	}

	log.Println("[RUN] files written")

	// 4 Build run command (must include < input.txt)
	runCmd := languageConfig.Cmd(names)
	log.Println("[RUN] command:", runCmd)

	log.Println("[RUN] using container:", containerID)

	// 6 Translate HOST path → CONTAINER path
	// IMPORTANT: container must mount /tmp/runner → /sandbox
	containerJobPath := filepath.Join("/sandbox", filepath.Base(basePath))

	log.Println("[RUN] Container Job Path:", basePath, containerJobPath)

	// 7 Create docker exec
	execCmd := []string{
		"sh", "-c",
		fmt.Sprintf("cd %s && %s", containerJobPath, runCmd),
	}

	execCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	execResp, err := d.client.Client.ContainerExecCreate(execCtx, containerID, container.ExecOptions{
		Cmd:          execCmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  false,
	})
	if err != nil {
		log.Println("[RUN] exec create failed:", err)
		return "", "", sandbox.ErrSandboxError, "Exec create failed"
	}

	attach, err := d.client.Client.ContainerExecAttach(execCtx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		log.Println("[RUN] exec attach failed:", err)
		return "", "", sandbox.ErrSandboxError, "Exec attach failed"
	}
	defer attach.Close()

	// 8 Read stdout / stderr from EXEC
	var stdoutBuf, stderrBuf bytes.Buffer

	done := make(chan error, 1)
	go func() {
		_, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attach.Reader)
		done <- err
	}()

	select {
	case <-execCtx.Done():
		log.Println("[RUN] TLE reached")
		return "", "", sandbox.ErrTLE, sandbox.MsgTLE

	case err := <-done:
		if err != nil {
			log.Println("[RUN] exec output read failed:", err)
			return "", "", sandbox.ErrSandboxError, "Failed reading exec output"
		}
	}

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	log.Println("[RUN] stdout:", stdout)
	log.Println("[RUN] stderr:", stderr)

	// 9 Error classification
	errType, errMsg := DetectError(language, stdout, stderr)
	if errType != "" {
		return stdout, stderr, errType, errMsg
	}

	return stdout, stderr, "", ""
}
