package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
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

func (s *OptionsService) ListFiles(directory string, recursive bool) ([]Option, error) {
	rootPath, ok := s.config.AllowedDirectories[directory]
	if !ok {
		return nil, fmt.Errorf("directory not allowed: %s", directory)
	}

	var files []Option

	appendFile := func(label, name string) {
		if !strings.HasSuffix(name, ".log") {
			return
		}

		files = append(files, Option{
			Label: strings.TrimSuffix(label, ".log"),
			Value: strings.TrimSuffix(name, ".log"),
		})
	}

	if !recursive {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, s.directoryError(directory, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			appendFile(entry.Name(), entry.Name())
		}

		return files, nil
	}

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil
		}

		appendFile(rel, d.Name())
		return nil
	})
	if err != nil {
		return nil, s.directoryError(directory, err)
	}

	return files, nil
}

func (s *OptionsService) ListContainers() ([]Option, error) {
	var containers []string

	switch s.config.DockerServicesMode {
	case "auto":
		c, err := s.listDockerContainers()
		if err != nil {
			return nil, err
		}
		containers = c

	case "manual":
		containers = s.config.DockerServices

	default:
		return nil, fmt.Errorf(
			"invalid docker services mode: %s",
			s.config.DockerServicesMode,
		)
	}

	options := make([]Option, 0, len(containers))
	for _, c := range containers {
		options = append(options, Option{
			Label: c,
			Value: c,
		})
	}

	return options, nil
}

func (s *OptionsService) listDockerContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		s.logger.Error("failed to list docker containers", logger.Field{Key: "error", Value: err.Error()})
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

func (s *OptionsService) directoryError(directory string, err error) error {
	s.logger.Error(
		"failed to list files in directory",
		logger.Field{Key: "directory", Value: directory},
		logger.Field{Key: "error", Value: err.Error()},
	)

	return fmt.Errorf("failed to read directory: %w", err)
}
