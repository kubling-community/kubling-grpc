package support

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	KublingCliImage = "kubling/kubling-cli:latest"

	ServerKeyStore = "server.ks"
	ClientKeyStore = "client.ks"

	ServerPassword = "myserverpass"
	ClientPassword = "myclientpass"

	BundleName = "test-descriptor-bundle.zip"
)

func PrepareScenario(
	scenarioDir string,
) error {

	if err := generateCertificates(scenarioDir); err != nil {
		return err
	}

	if err := generateBundle(scenarioDir); err != nil {
		return err
	}

	return chownGeneratedFiles(
		scenarioDir,
		ServerKeyStore,
		ClientKeyStore,
		BundleName,
	)

}

func generateBundle(
	scenarioDir string,
) error {

	return dockerRun(
		"run",
		"--rm",

		"-v",
		fmt.Sprintf("%s:/kbl", filepath.Clean(scenarioDir)),

		KublingCliImage,

		"bundle",
		"genmod",

		"/kbl/descriptor",

		"-o",
		"/kbl/"+BundleName,
	)

}

func generateCertificates(
	scenarioDir string,
) error {

	return dockerRun(
		"run",
		"--rm",

		"-v",
		fmt.Sprintf("%s:/kbl", filepath.Clean(scenarioDir)),

		KublingCliImage,

		"cert",
		"create",

		"-s",
		"/kbl/"+ServerKeyStore,

		"-c",
		"/kbl/"+ClientKeyStore,

		"-x",
		ServerPassword,

		"-p",
		ClientPassword,

		"-o",
	)

}

func dockerRun(
	args ...string,
) error {

	cmd := exec.Command(
		"docker",
		args...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func chownGeneratedFiles(
	baseDir string,
	files ...string,
) error {

	args := []string{
		"run",
		"--rm",

		"-v",
		fmt.Sprintf("%s:/work", filepath.Clean(baseDir)),

		"alpine",

		"chown",
		fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}

	for _, file := range files {
		args = append(
			args,
			filepath.ToSlash(filepath.Join("/work", file)),
		)
	}

	return dockerRun(args...)
}
