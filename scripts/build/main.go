package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("find", "./cmd", "-type", "f", "-name", "main.go")
	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	mainFiles := strings.Split(string(stdout), "\n")

	cmd = exec.Command("mkdir", "-p", "build")
	_, err = cmd.Output()
	if err != nil {
		panic(err)
	}

	hasError := false

	for _, file := range mainFiles {
		if len(file) < 1 {
			continue
		}

		err = buildLambda(file)
		if err != nil {
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}

func buildLambda(mainFile string) error {
	inputDir := getInputDirectory(mainFile)
	outPath := getOutputPath(mainFile)

	fmt.Printf("Build %s\n", inputDir)
	cmd := exec.Command("go", "build", "-o", outPath, "-ldflags=-w -s", inputDir)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux")
	cmd.Env = append(cmd.Env, "GOARCH=arm64")
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()
	if err != nil {
		fmt.Printf("\n")
		return err
	}
	fmt.Printf("OK    %s\n\n", outPath)
	return nil
}

func getInputDirectory(mainFile string) string {
	return strings.Replace(mainFile, "/main.go", "", 1)
}

// getOutputPath flattens the directory structure replacing `/` with `-` and sets the correct output directory
func getOutputPath(mainFile string) string {
	outDir := strings.Replace(mainFile, "/main.go", "", 1)
	outDir = strings.Replace(outDir, "./cmd/", "", 1)
	outDir = strings.ReplaceAll(outDir, "/", "-")

	return fmt.Sprintf("build/%s/bootstrap", outDir)
}
