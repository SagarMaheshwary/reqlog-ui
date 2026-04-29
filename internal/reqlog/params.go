package reqlog

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

var defaultLimit = 50

type CMDArgs struct {
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

func ParseParams(c *gin.Context, maxLines int) (*CMDArgs, error) {
	recursive := c.DefaultQuery("recursive", "true") != "false"

	service, err := validateService(c.Query("service"))
	if err != nil {
		return nil, err
	}

	key, err := validateKey(c.Query("key"))
	if err != nil {
		return nil, err
	}

	since, err := validateSince(c.Query("since"))
	if err != nil {
		return nil, err
	}

	source := c.DefaultQuery("source", "file")
	if source != "file" && source != "docker" {
		source = "file"
	}

	dir := ""
	if source == "file" {
		dir, err = validateDir(c.DefaultQuery("dir", "./logs"))
		if err != nil {
			return nil, err
		}
	}

	return &CMDArgs{
		SearchValue: validateQuery(c.Query("q")),
		Dir:         dir,
		IgnoreCase:  parseBool(c.Query("ignore_case")),
		Limit: validateLimit(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)),
			defaultLimit,
			maxLines,
		),
		JSON:      parseBool(c.Query("json")),
		Key:       key,
		Since:     since,
		Recursive: recursive,
		Service:   service,
		Source:    source,
	}, nil
}

func BuildArgs(p *CMDArgs, follow bool) []string {
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

func parseBool(v string) bool {
	return v == "true" || v == "1"
}
