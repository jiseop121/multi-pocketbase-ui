package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

var ErrExitRequested = errors.New("exit requested")

type REPLConfig struct {
	Stdin       io.Reader
	Stdout      io.Writer
	HistoryFile string
	Execute     func(line string) error
	Complete    func(line string) []string
}

func RunREPL(ctx context.Context, stdin io.Reader, stdout io.Writer, execute func(line string) error) error {
	return RunREPLWithConfig(ctx, REPLConfig{
		Stdin:   stdin,
		Stdout:  stdout,
		Execute: execute,
	})
}

func RunREPLWithConfig(ctx context.Context, cfg REPLConfig) error {
	if canUseReadline(cfg.Stdin, cfg.Stdout) {
		return runReadlineREPL(ctx, cfg)
	}
	return runScannerREPL(ctx, cfg)
}

func IsTTY(stdin io.Reader, stdout io.Writer) bool {
	inFile, inOK := stdin.(*os.File)
	outFile, outOK := stdout.(*os.File)
	if !inOK || !outOK {
		return false
	}
	return term.IsTerminal(int(inFile.Fd())) && term.IsTerminal(int(outFile.Fd()))
}

func canUseReadline(stdin io.Reader, stdout io.Writer) bool {
	return IsTTY(stdin, stdout)
}

func runScannerREPL(ctx context.Context, cfg REPLConfig) error {
	scanner := bufio.NewScanner(cfg.Stdin)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err := fmt.Fprint(cfg.Stdout, "pbviewer> "); err != nil {
			return err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
			return nil
		}

		err := cfg.Execute(line)
		if errors.Is(err, ErrExitRequested) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func runReadlineREPL(ctx context.Context, cfg REPLConfig) error {
	inFile := cfg.Stdin.(*os.File)
	outFile := cfg.Stdout.(*os.File)
	if strings.TrimSpace(cfg.HistoryFile) != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.HistoryFile), 0o755); err != nil {
			return err
		}
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "pbviewer> ",
		Stdin:           inFile,
		Stdout:          outFile,
		Stderr:          outFile,
		AutoComplete:    dynamicCompleter{fn: cfg.Complete},
		HistoryFile:     cfg.HistoryFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if strings.TrimSpace(line) == "" {
				continue
			}
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
			return nil
		}

		execErr := cfg.Execute(line)
		if errors.Is(execErr, ErrExitRequested) {
			return nil
		}
		if execErr != nil {
			return execErr
		}
	}
}

type dynamicCompleter struct {
	fn func(line string) []string
}

func (c dynamicCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if c.fn == nil {
		return nil, 0
	}
	if pos < 0 || pos > len(line) {
		pos = len(line)
	}
	prefixLine := string(line[:pos])
	wordPrefix := extractWordPrefix(prefixLine)
	candidates := c.fn(prefixLine)
	if len(candidates) == 0 {
		return nil, 0
	}
	out := make([][]rune, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, wordPrefix) {
			out = append(out, []rune(candidate))
		}
	}
	if len(out) == 0 {
		return nil, 0
	}
	return out, len([]rune(wordPrefix))
}

func extractWordPrefix(line string) string {
	runes := []rune(line)
	idx := len(runes) - 1
	for idx >= 0 {
		if unicode.IsSpace(runes[idx]) {
			break
		}
		idx--
	}
	if idx == len(runes)-1 {
		return ""
	}
	return string(runes[idx+1:])
}
