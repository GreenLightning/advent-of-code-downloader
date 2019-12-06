package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// This is a build script that produces the assets for a release.
// Use the following command to execute the script:
// go run build.go

func main() {
	folderName := "build"

	check(os.MkdirAll(folderName, 0777))
	check(os.Chdir("aocdl"))

	build("windows", "amd64", "", "aocdl.exe", "aocdl-windows.zip", folderName)
	build("darwin", "amd64", "", "aocdl", "aocdl-macos.zip", folderName)
	build("linux", "amd64", "", "aocdl", "aocdl-linux-amd64.zip", folderName)
	build("linux", "arm", "6", "aocdl", "aocdl-armv6.zip", folderName)
}

func build(goos, goarch, goarm string, binaryName, packageName, folderName string) {
	fullBinaryName := fmt.Sprintf("../%s/%s", folderName, binaryName)
	fullPackageName := fmt.Sprintf("../%s/%s", folderName, packageName)

	cmd := exec.Command("go", "build", "-o", fullBinaryName)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", goos))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", goarch))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOARM=%s", goarm))

	check(cmd.Run())
	check(createZip(fullBinaryName, fullPackageName))
	check(os.Remove(fullBinaryName))
}

func createZip(fullBinaryName, fullPackageName string) error {
	zipFile, err := os.Create(fullPackageName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	binaryFile, err := os.Open(fullBinaryName)
	if err != nil {
		return err
	}
	defer binaryFile.Close()

	info, err := binaryFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate
	header.SetMode(0755)

	fileWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileWriter, binaryFile)
	return err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
