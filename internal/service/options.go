package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
)

type OptionsService struct {
	config *config.Reqlog
	logger logger.Logger
}

type OptionsServiceOpts struct {
	Logger logger.Logger
	Config *config.Reqlog
}

func NewOptionsService(opts *OptionsServiceOpts) *OptionsService {
	return &OptionsService{
		config: opts.Config,
		logger: opts.Logger,
	}
}

func (s *OptionsService) ListDirectories() ([]string, error) {
	dirs := make([]string, 0, len(s.config.AllowedDirectories))
	for k := range s.config.AllowedDirectories {
		dirs = append(dirs, k)
	}
	return dirs, nil
}

func (s *OptionsService) ListFiles(directory string) ([]string, error) {
	path, ok := s.config.AllowedDirectories[directory]
	if !ok {
		return nil, fmt.Errorf("directory not allowed: %s", path)
	}

	files := make([]string, 0)

	dirFiles, err := os.ReadDir(path)
	if err != nil {
		s.logger.Error("failed to list files in directory", logger.Field{Key: "directory", Value: directory}, logger.Field{Key: "error", Value: err})
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range dirFiles {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".log") {
			continue
		}
		logFile, _ := strings.CutSuffix(file.Name(), ".log")
		files = append(files, logFile)
	}

	return files, nil
}

func (s *OptionsService) ListContainers() ([]string, error) {
	switch s.config.DockerServicesMode {
	case "auto":
		return s.listDockerContainers()
	case "manual":
		return s.config.DockerServices, nil
	default:
		return nil, fmt.Errorf("invalid docker services mode: %s", s.config.DockerServicesMode)
	}
}

func (s *OptionsService) listDockerContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		s.logger.Error("failed to list docker containers", logger.Field{Key: "error", Value: err})
		return nil, fmt.Errorf("failed to list docker containers: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	s.logger.Info("found docker containers", logger.Field{Key: "count", Value: len(lines)})

	var containers []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			containers = append(containers, l)
		}
	}

	return containers, nil
}
