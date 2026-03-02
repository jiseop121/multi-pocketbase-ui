package app

import (
	"io"
	"strings"
)

type ExecMode string

const (
	ModeREPL       ExecMode = "repl"
	ModeOneShot    ExecMode = "one-shot"
	ModeScript     ExecMode = "script"
	ModeUIReserved ExecMode = "ui-reserved"
)

type RunConfig struct {
	UIEnabled   bool
	CommandText string
	ScriptPath  string
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
}

func ParseRunConfig(args []string, stdin io.Reader, stdout, stderr io.Writer) (RunConfig, error) {
	cfg := RunConfig{Stdin: stdin, Stdout: stdout, Stderr: stderr}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-ui":
			cfg.UIEnabled = true
		case "-c":
			if i+1 >= len(args) {
				return cfg, NewInvalidArgsError("Missing command text for `-c`.", "Example: pbviewer -c \"version\"")
			}
			if strings.TrimSpace(args[i+1]) == "" {
				return cfg, NewInvalidArgsError("Command text for `-c` cannot be empty.", "Example: pbviewer -c \"version\"")
			}
			cfg.CommandText = args[i+1]
			i++
		case "-h", "--help", "help":
			cfg.CommandText = "help"
		case "version", "--version":
			if cfg.CommandText == "" {
				cfg.CommandText = "version"
			}
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return cfg, NewInvalidArgsError("Unknown option `"+arg+"`.", "Run `pbviewer -c \"help\"` to see available commands.")
			}
			if cfg.ScriptPath != "" {
				return cfg, NewInvalidArgsError("Only one script file path can be provided.", "Use: pbviewer <script-file>")
			}
			cfg.ScriptPath = arg
		}
	}

	if err := ValidateRunConfig(cfg); err != nil {
		return RunConfig{}, err
	}
	return cfg, nil
}

func ValidateRunConfig(cfg RunConfig) error {
	if cfg.CommandText != "" && cfg.ScriptPath != "" {
		return NewInvalidArgsError("Cannot use `-c` and script file path together.", "Choose one mode: `pbviewer -c \"...\"` or `pbviewer <script-file>`")
	}
	if cfg.UIEnabled && (cfg.CommandText != "" || cfg.ScriptPath != "") {
		return NewInvalidArgsError("`-ui` cannot be used with `-c` or script mode.", "Run `pbviewer -ui` alone.")
	}
	return nil
}

func ResolveMode(cfg RunConfig) ExecMode {
	if cfg.UIEnabled {
		return ModeUIReserved
	}
	if cfg.CommandText != "" {
		return ModeOneShot
	}
	if cfg.ScriptPath != "" {
		return ModeScript
	}
	return ModeREPL
}
