package service

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/limiter"
	"github.com/sagarmaheshwary/reqlog-ui/internal/reqlog"
)

type ReqlogService interface {
	Run(ctx context.Context, params *reqlog.CMDArgs) ([]string, error)
	Stream(ctx context.Context, params *reqlog.CMDArgs, out chan<- string) (<-chan error, error)
}

type reqlogService struct {
	config        *config.Reqlog
	searchLimiter *limiter.Limiter
	streamLimiter *limiter.Limiter
}

type ReqlogServiceOpts struct {
	Config *config.Reqlog
}

func NewReqlogService(opts ReqlogServiceOpts) ReqlogService {
	return &reqlogService{
		config:        opts.Config,
		searchLimiter: limiter.New(opts.Config.SearchConcurrency),
		streamLimiter: limiter.New(opts.Config.StreamConcurrency),
	}
}

func (s *reqlogService) Run(ctx context.Context, params *reqlog.CMDArgs) ([]string, error) {
	if !s.searchLimiter.TryAcquire() {
		return nil, &TooManyRequestsError{
			Message: "System is busy, try again shortly",
			Active:  s.searchLimiter.Active(),
			Limit:   s.searchLimiter.Limit(),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.ExecutionTimeout)
	defer cancel()

	args := reqlog.BuildArgs(params, false)
	cmd := exec.CommandContext(ctx, s.config.BinaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.searchLimiter.Release()
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		s.searchLimiter.Release()
		return nil, err
	}

	type result struct {
		lines []string
		err   error
	}

	resCh := make(chan result, 1)

	go func() {
		defer s.searchLimiter.Release()

		reader := bufio.NewReader(stdout)
		lines := make([]string, 0, 50)

		for {
			line, err := reader.ReadString('\n')

			if len(lines) < s.config.MaxLines && len(line) > 0 {
				lines = append(lines, strings.TrimRight(line, "\r\n"))
			}

			if err != nil {
				if err != io.EOF {
					resCh <- result{nil, err}
					return
				}
				break
			}
		}
		err := cmd.Wait()
		resCh <- result{lines, err}
	}()

	select {
	case res := <-resCh:
		if res.err != nil {
			return nil, res.err
		}
		return res.lines, nil

	case <-ctx.Done():
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, ctx.Err()
	}
}

func (s *reqlogService) Stream(ctx context.Context, params *reqlog.CMDArgs, out chan<- string) (<-chan error, error) {
	if !s.streamLimiter.TryAcquire() {
		return nil, &TooManyRequestsError{
			Message: "System is busy, try again shortly",
			Active:  s.streamLimiter.Active(),
			Limit:   s.streamLimiter.Limit(),
		}
	}

	args := reqlog.BuildArgs(params, true)
	cmd := exec.CommandContext(ctx, s.config.BinaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.streamLimiter.Release()
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		s.streamLimiter.Release()
		return nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		defer s.streamLimiter.Release()
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
