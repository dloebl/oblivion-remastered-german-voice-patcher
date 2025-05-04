package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: create-pak <directory>")
		os.Exit(1)
	}
	// Get the input directory from command line arguments
	inputDir := os.Args[1]
	// Get the working directory
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	// Load bin path from environment
	unrealBinDir := os.Getenv("UNREAL_BIN_DIR")
	if unrealBinDir == "" {
		// Error if the environment variable is not set
		log.Fatal("Error: UNREAL_BIN_DIR environment variable is not set")
	}

	// Define file paths
	fileList := filepath.Join(workingDir, "tmp", "filelist.txt")
	outPak := filepath.Join(workingDir, "ModFiles", "Content", "Paks", "~mods", "german-voices-oblivion-remastered-voxmeld_v0.4.2_P.pak")

	// Create ModFiles directory structure if it doesn't exist
	modDir := filepath.Dir(outPak)
	os.MkdirAll(modDir, 0755)

	// Create file list
	err = createFileList(inputDir, fileList)
	if err != nil {
		fmt.Printf("Error creating file list: %v\n", err)
		os.Exit(1)
	}

	// Run UnrealPak
	unrealPakExe := filepath.Join(unrealBinDir, "UnrealPak.exe")
	cmd := exec.Command(
		unrealPakExe,
		outPak,
		"-create="+fileList,
		"-compress",
		"-compressionformats=Oodle",
		"-compressmethod=Kraken",
		"-compresslevel=4",
	)
	cmd.Dir = unrealBinDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Creating PAK file...")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running UnrealPak: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PAK file created successfully.")
}

// createFileList creates a file list for UnrealPak from the specified directory
func createFileList(inputDir, outputFile string) error {
	// Open output file
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Walk through all files in the input directory
	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Calculate relative path
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return err
		}
		// Get absolute path for the source file
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		// Format the line for UnrealPak using absolute path for the source file
		line := fmt.Sprintf("%q \"../../../OblivionRemastered/%s\"\n", absPath, relPath)
		_, err = writer.WriteString(line)
		return err
	})

	if err != nil {
		return err
	}

	return writer.Flush()
}
