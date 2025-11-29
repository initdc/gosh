package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func builtinLs(args []string) {
	showDotfiles := false
	filteredArgs := []string{}

	// Parse arguments
	for _, arg := range args {
		if arg == "-a" || arg == "--all" {
			showDotfiles = true
		} else if !strings.HasPrefix(arg, "-") {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	dirPath := "."
	if len(filteredArgs) > 0 {
		dirPath = filteredArgs[0]
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("ls: cannot access '%s': No such file or directory\n", dirPath)
		return
	}

	// Collect file names
	var entries []string
	for _, file := range files {
		entries = append(entries, file.Name())
	}

	// Sort entries
	sort.Strings(entries)

	// Print entries
	for _, entry := range entries {
		if showDotfiles || !strings.HasPrefix(entry, ".") {
			fmt.Println(entry)
		}
	}
}

func builtinCd(args []string) {
	if len(args) != 1 {
		fmt.Println("cd: expected one argument")
		return
	}

	dir := args[0]
	err := os.Chdir(dir)
	if err != nil {
		fmt.Printf("cd: no such file or directory: %s\n", dir)
	}
}

func builtinExit(args []string) {
	if len(args) > 1 {
		fmt.Println("exit: too many arguments")
		return
	}

	exitCode := 0
	if len(args) == 1 {
		code, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("exit: numeric argument required")
			exitCode = 1
		} else {
			exitCode = code
		}
	}

	os.Exit(exitCode)
}

func builtinHelp(args []string) {
	_ = args
	fmt.Println("Available built-in commands:")
	fmt.Println("  cd [directory]   Change the current directory to 'directory'.")
	fmt.Println("  exit [n]        Exit the shell with a status of 'n'.")
	fmt.Println("  help            Display this help message.")
}

func isBuiltinCommand(command string) bool {
	switch command {
	case "ls", "cd", "exit", "help":
		return true
	default:
		return false
	}
}

func isGoPrintCommand(command string) bool {
	return strings.HasPrefix(command, "fmt.")
}

func executeBuiltin(command string, args []string) {
	switch command {
	case "ls":
		builtinLs(args)
	case "cd":
		builtinCd(args)
	case "exit":
		builtinExit(args)
	case "help":
		builtinHelp(args)
	}
}

func goEval(code string) {
	// Create a temporary Go file
	tmpFile, err := os.CreateTemp("", "eval_*.go")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Filter code to only include lines starting with fmt.
	var filteredCode strings.Builder
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "fmt.") {
			filteredCode.WriteString(line)
			filteredCode.WriteString("\n")
		}
	}

	// Write the code to the file
	_, err = fmt.Fprintf(tmpFile, `package main

import (
	"fmt"
)

func main() {
	%s
}
`, filteredCode.String())
	if err != nil {
		fmt.Printf("Error writing to temp file: %v\n", err)
		return
	}

	// Run the code
	cmd := exec.Command("go", "run", tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running code: %v\n", err)
	}
}

func evaluateGoPrint(input string) {
	goEval(input)
}

func runCommand(input string) {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	if isBuiltinCommand(command) {
		executeBuiltin(command, args)
	} else if isGoPrintCommand(command) {
		evaluateGoPrint(input)
	} else {
		// Execute external command
		cmd := exec.Command(command, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("%s: command not found\n", command)
		}
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			// Check for EOF (Ctrl+D) and exit gracefully
			if err.Error() == "EOF" {
				fmt.Println("")
				os.Exit(0)
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		runCommand(input)
	}
}
