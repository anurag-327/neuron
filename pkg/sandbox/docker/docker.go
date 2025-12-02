package docker

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerClient struct {
	client *client.Client
}

func GetDockerClient() (sandbox.Runner, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &DockerClient{client: cli}, nil
}

//
// -----------------------------------------------------------------------------
//  Run: Executes code in Docker Sandbox
// -----------------------------------------------------------------------------
//  Returns:
//    stdout, stderr, errType, errMessage
// -----------------------------------------------------------------------------

func (d *DockerClient) Run(ctx context.Context, basePathString, code, input, language string) (string, string, sandbox.SandboxError, sandbox.SandboxErrorMessage) {

	// 1) Create job directory
	projectRoot, _ := os.Getwd()
	basePath := filepath.Join(projectRoot, basePathString)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return "", "", sandbox.ErrInternalError, "Failed to create job directory"
	}

	// 2) Load language config + build filenames
	languageConfig, err := GetLanguageConfig(language)
	if err != nil {
		return "", "", sandbox.ErrInternalError, "Unsupported language"
	}

	names := BuildFileNames(basePath, languageConfig)

	// 3) Write code + input
	if err := util.WriteContentToFile(names.PathFull, []byte(code), 0644); err != nil {
		return "", "", sandbox.ErrInternalError, "Failed to write code file"
	}

	inputPath := filepath.Join(basePath, "input.txt")
	if err := util.WriteContentToFile(inputPath, []byte(input), 0644); err != nil {
		return "", "", sandbox.ErrInternalError, "Failed to write input file"
	}

	// 4) Get image + build command
	imageName := languageConfig.DockerImage
	command := languageConfig.Cmd(names)

	// 5) Pull image
	reader, err := d.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", "", sandbox.ErrSandboxError, "Failed to pull sandbox image"
	}
	io.Copy(io.Discard, reader)
	reader.Close()

	// 6) Build the actual shell command
	runCmd := command

	absPath, _ := filepath.Abs(basePath)

	// 7) Create container
	resp, err := d.client.ContainerCreate(
		ctx,
		&container.Config{
			Image:       imageName,
			Cmd:         []string{"sh", "-c", runCmd},
			WorkingDir:  "/app",
			AttachStdin: true,
			OpenStdin:   true,
			StdinOnce:   true,
			Env:         []string{"GOCACHE=/tmp/go-cache"},
		},
		&container.HostConfig{
			AutoRemove: false,
			Mounts: []mount.Mount{
				{Type: mount.TypeBind, Source: absPath, Target: "/app"},
				{Type: mount.TypeTmpfs, Target: "/tmp"},
			},
			ReadonlyRootfs: false,
			SecurityOpt:    []string{"no-new-privileges:true"},
			NetworkMode:    "none",
			Resources: container.Resources{
				Memory:    256 * 1024 * 1024,
				NanoCPUs:  1_000_000_000,
				PidsLimit: func(i int64) *int64 { return &i }(50),
			},
		},
		nil, nil, "",
	)

	if err != nil {
		return "", "", sandbox.ErrSandboxError, "Failed to create sandbox container"
	}

	// 8) Attach to container stdin
	attachConn, err := d.client.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return "", "", sandbox.ErrSandboxError, "Failed to attach container stdin"
	}

	// Stream input into container's stdin
	if input != "" {
		io.Copy(attachConn.Conn, strings.NewReader(input))
	}
	attachConn.CloseWrite()

	// 9) Start container
	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", sandbox.ErrSandboxError, "Failed to start sandbox container"
	}

	// 10) TLE handling
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	statusCh, errCh := d.client.ContainerWait(timeoutCtx, resp.ID, container.WaitConditionNotRunning)

	select {
	case <-timeoutCtx.Done():
		_ = d.client.ContainerKill(context.Background(), resp.ID, "SIGKILL")
		_ = d.client.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})
		return "", "", sandbox.ErrTLE, sandbox.MsgTLE

	case err := <-errCh:
		if err != nil {
			return "", "", sandbox.ErrSandboxError, "Sandbox container wait failure"
		}

	case <-statusCh:
	}

	// 11) Collect logs — unchanged
	logs, err := d.client.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", "", sandbox.ErrSandboxError, "Failed to read sandbox logs"
	}
	defer logs.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	stdcopy.StdCopy(&stdoutBuf, &stderrBuf, logs)

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	// 12) Cleanup container
	_ = d.client.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})

	// 13) Error classification
	errType, errMsg := DetectError(language, stdout, stderr)
	if errType != "" {
		return stdout, stderr, errType, errMsg
	}

	if stderr != "" {
		return stdout, stderr, sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
	}

	return stdout, stderr, "", ""
}
