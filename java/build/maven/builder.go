package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const baseSpace = "/root/src"

type Builder struct {
	// 用户提供参数, 通过环境变量传入
	GitCloneURL string
	GitRef      string
	Goals       string
	PomPath     string

	HubBinRepo string
	HubUser    string
	HubToken   string
	BinTag     string
	BinPath    string

	projectName string
}

// NewBuilder is
func NewBuilder(envs map[string]string) (*Builder, error) {
	b := &Builder{}

	if envs["GIT_CLONE_URL"] == "" {
		return nil, fmt.Errorf("envionment variables GIT_CLONE_URL is required")
	}

	b.GitCloneURL = envs["GIT_CLONE_URL"]
	s := strings.TrimSuffix(strings.TrimSuffix(b.GitCloneURL, "/"), ".git")
	b.projectName = s[strings.LastIndex(s, "/")+1:]

	if b.GitRef = envs["GIT_REF"]; b.GitRef == "" {
		b.GitRef = "master"
	}

	b.Goals = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(envs["GOALS"]), "mvn "))

	if b.PomPath = envs["POM_PATH"]; b.PomPath == "" {
		b.PomPath = "./pom.xml"
	}

	b.HubBinRepo = envs["HUB_BIN_REPO"]
	b.HubUser = envs["HUB_USER"]
	b.HubToken = envs["HUB_TOKEN"]
	b.BinPath = envs["BIN_PATH"]
	b.BinTag = envs["BIN_TAG"]
	if b.BinTag == "" {
		b.BinTag = "latest"
	}

	return b, nil
}

func (b *Builder) run() error {
	if err := os.Chdir(baseSpace); err != nil {
		return fmt.Errorf("Chdir to baseSpace(%s) failed:%v", baseSpace, err)
	}

	if err := b.gitPull(); err != nil {
		return err
	}

	if err := b.gitReset(); err != nil {
		return err
	}

	if err := b.build(); err != nil {
		return err
	}

	if err := b.handleArtifacts(); err != nil {
		return err
	}
	// err = b.doPush(b.Image)
	// if err != nil {
	// 	return
	// }
	return nil
}

func (b *Builder) build() error {
	var command = strings.Fields(b.Goals)

	if len(command) == 0 {
		command = append(command, "mvn", "package")
	}

	if command[0] != "mvn" {
		command = append([]string{"mvn"}, command...)
	}

	command = append(command, "-f", b.PomPath)

	cwd, _ := os.Getwd()
	if _, err := (CMD{command, filepath.Join(cwd, b.projectName)}).Run(); err != nil {
		fmt.Println("Run mvn goals failed:", err)
		return err
	}
	fmt.Println("Run mvn goals succeded.")
	return nil
}

func (b *Builder) handleArtifacts() error {
	cwd, _ := os.Getwd()
	targetPath := filepath.Join(cwd, b.projectName, "target")

	command := []string{"find", "./", "-name", "*.jar", "-o", "-name", "*.war"}
	output, err := (CMD{command, targetPath}).Run()
	if err != nil {
		fmt.Println("Run find artifacts failed:", err)
		return err
	}
	output = strings.TrimSpace(output)
	if len(output) == 0 {
		return errors.New("no artifact")
	}

	artifactsSlice := strings.Split(output, "\n")
	fmt.Printf("[JOB_OUT] ARTIFACTS = %s\n", strings.Join(artifactsSlice, ";"))

	if b.HubBinRepo == "" {
		fmt.Println("HUB_BIN_REPO is empty, no need upload artifacts")
		return nil
	}

	artifactsTar := fmt.Sprintf("%s.tar", b.projectName)
	command = []string{"tar", "-cf", artifactsTar}
	command = append(command, artifactsSlice...)
	if _, err := (CMD{command, targetPath}).Run(); err != nil {
		fmt.Println("Run tar artifacts failed:", err)
		return err
	}

	command = []string{
		"/.workflow/bin/thub", "push",
		fmt.Sprintf("--username=%s", b.HubUser), fmt.Sprintf("--password=%s", b.HubToken),
		fmt.Sprintf("--repo=%s", b.HubBinRepo),
		fmt.Sprintf("--localpath=%s", artifactsTar),
		fmt.Sprintf("--path=%s", filepath.Join(b.BinPath, artifactsTar)),
		fmt.Sprintf("--tag=%s", b.BinTag),
	}
	if _, err := (CMD{command, targetPath}).Run(); err != nil {
		fmt.Println("Run upload artifacts failed:", err)
		return err
	}

	// TODO
	fmt.Printf("[JOB_OUT] BIN_URL = %s\n", filepath.Join(b.HubBinRepo, b.BinPath, artifactsTar))
	fmt.Println("Run upload artifacts succeded.")
	return nil
}

func (b *Builder) gitPull() error {
	var command = []string{"git", "clone", "--recurse-submodules", b.GitCloneURL, b.projectName}
	if _, err := (CMD{Command: command}).Run(); err != nil {
		fmt.Println("Clone project failed:", err)
		return err
	}
	fmt.Println("Clone project", b.GitCloneURL, "succeded.")
	return nil
}

func (b *Builder) gitReset() error {
	cwd, _ := os.Getwd()
	var command = []string{"git", "reset", "--hard", b.GitRef}
	if _, err := (CMD{command, filepath.Join(cwd, b.projectName)}).Run(); err != nil {
		fmt.Println("Switch to commit", b.GitRef, "failed:", err)
		return err
	}
	fmt.Println("Switch to", b.GitRef, "succeded.")
	return nil
}

type CMD struct {
	Command []string // cmd with args
	WorkDir string
}

func (c CMD) Run() (string, error) {
	fmt.Println("Run CMD: ", strings.Join(c.Command, " "))

	cmd := exec.Command(c.Command[0], c.Command[1:]...)
	if c.WorkDir != "" {
		cmd.Dir = c.WorkDir
	}

	data, err := cmd.CombinedOutput()
	result := string(data)
	if len(result) > 0 {
		fmt.Println(result)
	}

	return result, err
}