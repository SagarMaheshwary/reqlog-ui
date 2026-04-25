package service

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/reqlog"
)

type ReqlogService interface {
	Run(ctx context.Context, params *reqlog.CMDArgs) ([]string, error)
	Stream(ctx context.Context, params *reqlog.CMDArgs, out chan<- string) (<-chan error, error)
}

type reqlogService struct {
	config *config.Reqlog
}

type ReqlogServiceOpts struct {
	Config *config.Reqlog
}

func NewReqlogService(opts ReqlogServiceOpts) ReqlogService {
	return &reqlogService{config: opts.Config}
}

func (s *reqlogService) Run(ctx context.Context, params *reqlog.CMDArgs) ([]string, error) {
	args := reqlog.BuildArgs(params, false)
	cmd := exec.CommandContext(ctx, s.config.BinaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	lines := make([]string, 0, 100)

	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				lines = append(lines, strings.TrimRight(line, "\r\n"))
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				done <- err
				break
			}
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			_ = cmd.Process.Kill()
			return nil, err
		}
		return lines, nil

	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return nil, ctx.Err()

	case <-time.After(s.config.ExecutionTimeout):
		_ = cmd.Process.Kill()
		return nil, errors.New("reqlog execution timeout")
	}
}

func (s *reqlogService) Stream(ctx context.Context, params *reqlog.CMDArgs, out chan<- string) (<-chan error, error) {
	args := reqlog.BuildArgs(params, true)
	cmd := exec.CommandContext(ctx, s.config.BinaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(out)

		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')

			if len(line) > 0 {
				select {
				case out <- strings.TrimRight(line, "\r\n"):
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}

			if err != nil {
				if err == io.EOF {
					errCh <- cmd.Wait()
					return
				}
				errCh <- err
				return
			}
		}
	}()

	return errCh, nil
}
