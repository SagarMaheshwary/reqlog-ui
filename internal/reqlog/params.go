package reqlog

import (
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

var defaultLimit = 25

type CMDArgs struct {
	SearchValue string
	Dir         string
	IgnoreCase  bool
	Limit       int
	Context     int
	Latest      bool
	Format      string
	Key         string
	Since       string
	Recursive   bool
	Service     string
	Source      string
	Output      string
	Fields      string
	Verbose     bool
}

func ParseParams(c *gin.Context, maxLines int, allowedDirs map[string]string) (*CMDArgs, error) {
	service, err := validateIdentifierList(c.Query("service"), "service")
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
		dir, err = validateDir(c.DefaultQuery("dir", "./logs"), allowedDirs)
		if err != nil {
			return nil, err
		}
	}

	format := c.DefaultQuery("format", "auto")
	if !slices.Contains([]string{"auto", "text", "json"}, format) {
		format = "auto"
	}

	output := c.DefaultQuery("output", "pretty")
	if !slices.Contains([]string{"pretty", "json"}, output) {
		output = "pretty"
	}

	fields, err := validateIdentifierList(c.Query("fields"), "fields")
	if err != nil {
		return nil, err
	}

	return &CMDArgs{
		SearchValue: validateQuery(c.Query("q")),
		Dir:         dir,
		IgnoreCase:  parseBool(c.Query("ignore_case")),
		Limit: validateLimit(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)),
			defaultLimit,
			maxLines,
		),
		Latest:    parseBool(c.Query("latest")),
		Key:       key,
		Since:     since,
		Recursive: parseBool(c.Query("recursive")),
		Service:   service,
		Source:    source,
		Format:    format,
		Output:    output,
		Context:   validateLimit(c.DefaultQuery("context", "0"), 0, maxLines),
		Fields:    fields,
		Verbose:   parseBool(c.Query("verbose")),
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
	if p.Latest {
		args = append(args, "--latest")
	}
	if p.Key != "" {
		args = append(args, "--key", p.Key)
	}
	if p.Since != "" {
		args = append(args, "--since", p.Since)
	}
	if p.Recursive {
		args = append(args, "--recursive")
	}
	if p.Service != "" {
		args = append(args, "--service", p.Service)
	}
	if p.Source != "" {
		args = append(args, "--source", p.Source)
	}
	if p.Format != "" {
		args = append(args, "--format", p.Format)
	}
	if p.Output != "" {
		args = append(args, "--output", p.Output)
	}
	if follow {
		args = append(args, "--follow")
	}
	if p.Context > 0 {
		args = append(args, "--context", strconv.Itoa(p.Context))
	}
	if p.Fields != "" {
		args = append(args, "--fields", p.Fields)
	}
	if p.Verbose {
		args = append(args, "--verbose")
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
