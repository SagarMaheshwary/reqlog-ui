package service

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"
)

type ReqlogParams struct {
	SearchValue string
	Dir         string
	IgnoreCase  bool
	Limit       int
	JSON        bool
	Key         string
	Since       string
	Recursive   bool
	Service     string
	Source      string
}

type ReqlogService interface {
	Run(ctx context.Context, params ReqlogParams) ([]string, error)
	Stream(ctx context.Context, params ReqlogParams, out chan<- string) error
}

type reqlogService struct {
	binaryPath string
}

type ReqlogServiceOpts struct {
	BinaryPath string
}

func NewReqlogService(opts ReqlogServiceOpts) ReqlogService {
	bp := opts.BinaryPath
	if bp == "" {
		bp = "reqlog"
	}
	return &reqlogService{binaryPath: bp}
}

func (s *reqlogService) Run(ctx context.Context, params ReqlogParams) ([]string, error) {
	args := buildArgs(params, false)
	cmd := exec.CommandContext(ctx, s.binaryPath, args...)

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			combined := string(out) + string(exitErr.Stderr)
			lines := splitLines(combined)
			return lines, nil
		}
		return nil, err
	}

	return splitLines(string(out)), nil
}

func (s *reqlogService) Stream(ctx context.Context, params ReqlogParams, out chan<- string) error {
	args := buildArgs(params, true)
	cmd := exec.CommandContext(ctx, s.binaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer close(out)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				return
			case out <- scanner.Text():
			}
		}
		cmd.Wait()
	}()

	return nil
}

func splitLines(s string) []string {
	var lines []string
	for l := range strings.SplitSeq(s, "\n") {
		lines = append(lines, l)
	}
	// Trim trailing empty line produced by a final \n.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func buildArgs(p ReqlogParams, follow bool) []string {
	var args []string

	if p.Dir != "" {
		args = append(args, "--dir", p.Dir)
	}
	if p.IgnoreCase {
		args = append(args, "--ignore-case")
	}
	if p.Limit > 0 {
		args = append(args, "--limit", strconv.Itoa(p.Limit))
	}
	if p.JSON {
		args = append(args, "--json")
	}
	if p.Key != "" {
		args = append(args, "--key", p.Key)
	}
	if p.Since != "" {
		args = append(args, "--since", p.Since)
	}
	if !p.Recursive {
		args = append(args, "--recursive=false")
	}
	if p.Service != "" {
		args = append(args, "--service", p.Service)
	}
	if p.Source != "" {
		args = append(args, "--source", p.Source)
	}
	if follow {
		args = append(args, "--follow")
	}
	// Search value goes last.
	if p.SearchValue != "" {
		args = append(args, p.SearchValue)
	}
	return args
}
