package main

import (
	"fmt"
	"io"
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
	outDir := getOutputDir(mainFile)
	outPath := fmt.Sprintf("%s/bootstrap", outDir)

	fmt.Printf("Build %s\n", inputDir)
	cmd := exec.Command("go", "build", "-o", outPath, "-trimpath", "-buildvcs=false", "-ldflags=-w -s", inputDir)
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
	size := float64(0)
	fi, err := os.Stat(outPath)
	if err == nil {
		size = float64(fi.Size()) / (1000 * 1000)
		fmt.Printf("OK    %s %.1fMB\n\n", outPath, size)
	} else {
		fmt.Printf("OK    %s\n\n", outPath)
	}

	cmd = exec.Command("find", inputDir, "-type", "f", "-name", "*.json")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	jsonFiles := strings.Split(string(stdout), "\n")

	for _, file := range jsonFiles {
		if file == "" {
			continue
		}
		dest := strings.Replace(file, inputDir, outDir, 1)
		_, err := copyFile(file, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func getInputDirectory(mainFile string) string {
	return strings.Replace(mainFile, "/main.go", "", 1)
}

// getOutputDir flattens the directory structure replacing `/` with `-` and sets the correct output directory
func getOutputDir(mainFile string) string {
	outDir := strings.Replace(mainFile, "/main.go", "", 1)
	outDir = strings.Replace(outDir, "./cmd/", "", 1)
	outDir = strings.ReplaceAll(outDir, "/", "-")

	return fmt.Sprintf("build/%s", outDir)
}
