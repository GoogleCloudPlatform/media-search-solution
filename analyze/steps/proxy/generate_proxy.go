package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultFfmpegArgs = "-analyzeduration 0 -probesize 5000000 -y -hide_banner -i %s -filter:v scale=w=%s:h=trunc(ow/a/2)*2 -f mp4 %s"
	TempFilePrefix    = "ffmpeg-output-"
	CommandSeparator  = " "
	FileCheckRetries  = 2
	FileCheckDelay    = 2 * time.Second
)

type RunConfig struct {
	InputFile    string
	MountPoint   string
	OutputFolder string
	TargetWidth  string
	CommandPath  string
	OutputFormat string
}

func main() {
	config, err := loadRunConfig()
	if err != nil {
		log.Fatal(err)
	}

	inputFileFullPath := config.MountPoint + "/" + config.InputFile

	for i := range FileCheckRetries {
		if _, err = os.Stat(inputFileFullPath); err == nil {
			break
		}
		log.Printf("waiting for file to appear: %s, attempt %d/%d", inputFileFullPath, i+1, FileCheckRetries)
		time.Sleep(FileCheckDelay)
	}

	if err != nil {
		log.Fatalf("input file %s not found after %d retries: %v", inputFileFullPath, FileCheckRetries, err)
	}

	file, err := os.Open(inputFileFullPath)
	if err != nil {
		log.Fatalf("error opening input file %s: %v", inputFileFullPath, err)
	}
	tempFile, _ := os.CreateTemp("", TempFilePrefix)

	args := fmt.Sprintf(DefaultFfmpegArgs, file.Name(), config.TargetWidth, tempFile.Name())
	cmd := exec.Command(config.CommandPath, strings.Split(args, CommandSeparator)...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("error running ffmpeg command: %v", err)
	}

	fileName := strings.SplitN(config.InputFile, "/", 2)[1]

	outputName := fileName
	if ext := filepath.Ext(fileName); ext != ".mp4" {
		outputName = strings.TrimSuffix(outputName, ext) + ".mp4"
	}

	outputFile := fmt.Sprintf("%s/%s/%s", config.MountPoint, config.OutputFolder, outputName)

	err = MoveFile(tempFile.Name(), outputFile)
	if err != nil {
		log.Fatalf("error moving file: %v", err)
	}

}

func loadRunConfig() (*RunConfig, error) {
	inputFile := os.Getenv("INPUT_FILE")
	if len(inputFile) == 0 {
		return &RunConfig{}, fmt.Errorf("no input file specified")
	}
	outputFolder := os.Getenv("OUTPUT_FOLDER")
	if len(outputFolder) == 0 {
		return &RunConfig{}, fmt.Errorf("no output folder specified")
	}

	mountPoint := getenv("MOUNT_POINT", "/mnt")
	targetWidth := getenv("OUTPUT_WIDTH", "240")
	commandPath := getenv("COMMAND_PATH", "bin/ffmpeg")
	outputFormat := getenv("OUTPUT_FORMAT", "mp4")

	return &RunConfig{
		InputFile:    inputFile,
		MountPoint:   mountPoint,
		OutputFolder: outputFolder,
		TargetWidth:  targetWidth,
		CommandPath:  commandPath,
		OutputFormat: outputFormat,
	}, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not open dest file: %v", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("could not copy to dest from source: %v", err)
	}

	inputFile.Close()

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("could not remove source file: %v", err)
	}
	return nil
}
