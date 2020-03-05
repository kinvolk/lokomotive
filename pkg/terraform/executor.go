// Copyright 2020 The Lokomotive Authors
// Copyright 2017 CoreOS, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/hpcloud/tail"
	"github.com/kardianos/osext"
	"github.com/shirou/gopsutil/process"
)

const (
	stateFileName  = "terraform.tfstate"
	tfVarsFileName = "terraform.tfvars"
	logsFolderName = "logs"

	logsFileSuffix = ".log"
	failFileSuffix = ".fail"

	requiredVersion = ">= 0.12, < 0.13"
)

// ErrBinaryNotFound denotes the fact that the Terraform binary could not be
// found on disk.
var ErrBinaryNotFound = errors.New(
	"Terraform not in executable's folder, cwd nor PATH",
)

// ExecutionStatus describes whether an execution succeeded, failed or is still
// in progress.
type ExecutionStatus string

const (
	// ExecutionStatusUnknown indicates that the status of execution is unknown.
	ExecutionStatusUnknown ExecutionStatus = "Unknown"
	// ExecutionStatusRunning indicates that the the execution is still in
	// process.
	ExecutionStatusRunning ExecutionStatus = "Running"
	// ExecutionStatusSuccess indicates that the execution succeeded.
	ExecutionStatusSuccess ExecutionStatus = "Success"
	// ExecutionStatusFailure indicates that the execution failed.
	ExecutionStatusFailure ExecutionStatus = "Failure"
)

// Executor enables calling Terraform from Go, across platforms, with any
// additional providers/provisioners that the currently executing binary
// exposes.
//
// The Terraform binary is expected to be in the executing binary's folder, in
// the current working directory or in the PATH.
// Each Executor runs in a temporary folder, so each Executor should only be
// used for one TF project.
//
// TODO: Ideally, we would use Terraform as a Go library, so we can monitor a
// hook and report the current state in real-time when
// Apply/Refresh/Destroy are used. While technically possible today, because
// Terraform currently hides the providers/provisioners list construction in
// their main package, it would require to reproduce a bunch of their logic,
// which is out of the scope of the first-version of the Executor. With a bit of
// efforts, we could actually even stop requiring having a Terraform binary
// altogether, by linking the builtin providers/provisioners to this particular
// binary and re-implemeting the routing here. Alternatively, we could
// contribute upstream to add a 'debug' flag that would enable a hook that would
// expose the live state to a file (or else).
type Executor struct {
	executionPath string
	binaryPath    string
	envVariables  map[string]string
	quiet         bool
}

// NewExecutor initializes a new Executor.
func NewExecutor(conf Config) (*Executor, error) {
	ex := new(Executor)
	ex.executionPath = conf.WorkingDir
	ex.quiet = conf.Quiet

	// Create the folder in which the executor, and its logs will be stored,
	// if not existing.
	os.MkdirAll(filepath.Join(ex.executionPath, logsFolderName), 0770)

	// Find the Terraform binary.
	out, err := tfBinaryPath()
	if err != nil {
		return nil, err
	}
	ex.binaryPath = out

	err = ex.checkVersion()
	if err != nil {
		return nil, err
	}

	return ex, nil
}

// Init() is a wrapper function that runs
// `terraform init`.
func (ex *Executor) Init() error {
	return ex.Execute("init")
}

// Apply() is a wrapper function that runs
// `terraform apply -auto-approve`.
func (ex *Executor) Apply() error {
	return ex.Execute("apply", "-auto-approve")
}

// Destroy() is a wrapper function that runs
// `terraform destroy -auto-approve`.
func (ex *Executor) Destroy() error {
	return ex.Execute("destroy", "-auto-approve")
}

// tailFile will indefinitely tail logs from the given file path, until
// given channel is closed.
func tailFile(path string, done chan struct{}, wg *sync.WaitGroup) {
	t, err := tail.TailFile(path, tail.Config{Follow: true})
	if err != nil {
		fmt.Printf("Unable to print logs from %s: %v\n", path, err)

		return
	}

	wg.Add(1)

	go func() {
		for line := range t.Lines {
			fmt.Println(line.Text)
		}

		wg.Done()
	}()

	<-done

	if err := t.Stop(); err != nil {
		fmt.Printf("Stopping printing logs from %s failed: %v\n", path, err)
	}

	wg.Done()
}

// Execute runs the given command and arguments against Terraform, and returns
// any errors that occur during the execution.
//
// An error is returned if the Terraform binary could not be found, or if the
// Terraform call itself failed, in which case, details can be found in the
// output.
func (ex *Executor) Execute(args ...string) error {
	pid, done, err := ex.ExecuteAsync(args...)
	if err != nil {
		return fmt.Errorf("failed executing Terraform command with arguments '%s' in directory %s: %w", strings.Join(args, " "), ex.WorkingDirectory(), err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	// Schedule waiting for Terraform execution to finish.
	go func() {
		<-done
		wg.Done()
	}()

	p := filepath.Join(ex.WorkingDirectory(), "logs", fmt.Sprintf("%d%s", pid, ".log"))

	// If we print output, schedule it as well.
	if !ex.quiet {
		wg.Add(1)

		go tailFile(p, done, &wg)
	}

	wg.Wait()

	s, err := ex.Status(pid)
	if err != nil {
		return fmt.Errorf("failed checking execution status: %w", err)
	}

	if s != ExecutionStatusSuccess {
		return fmt.Errorf("executing Terraform failed, check %s for details", p)
	}

	return nil
}

// LoadVars is a convenience function to load the tfvars file into memory
// as a JSON object.
func (ex *Executor) LoadVars() (map[string]interface{}, error) {
	filePath := filepath.Join(ex.WorkingDirectory(), tfVarsFileName)
	txt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var obj interface{}
	if err = json.Unmarshal([]byte(txt), &obj); err != nil {
		return nil, err
	}
	if data, ok := obj.(map[string]interface{}); ok {
		return data, nil
	}
	return nil, errors.New("Could not parse config as JSON object")
}

// ExecuteAsync runs the given command and arguments against Terraform, and returns
// an identifier that can be used to read the output of the process as it is
// executed and after.
//
// ExecuteAsync is non-blocking, and takes a lock in the execution path.
// Locking is handled by Terraform itself.
//
// An error is returned if the Terraform binary could not be found, or if the
// Terraform call itself failed, in which case, details can be found in the
// output.
func (ex *Executor) ExecuteAsync(args ...string) (int, chan struct{}, error) {
	cmd := ex.generateCommand(args...)
	rPipe, wPipe := io.Pipe()
	cmd.Stdout = wPipe
	cmd.Stderr = wPipe

	// Start Terraform.
	err := cmd.Start()
	if err != nil {
		// The process failed to start, we can't even save that it started since we
		// don't have a PID yet.
		return -1, nil, err
	}

	// Create a log file and pipe stdout/stderr to it.
	logFile, err := os.Create(ex.logPath(cmd.Process.Pid))
	if err != nil {
		return -1, nil, err
	}
	go io.Copy(logFile, rPipe)

	done := make(chan struct{})
	go func() {
		// Wait for the process to finish.
		if err := cmd.Wait(); err != nil {
			// The process did not end cleanly. Write the failure file.
			ioutil.WriteFile(ex.failPath(cmd.Process.Pid), []byte(err.Error()), 0660)
		}

		// Close descriptors.
		wPipe.Close()
		logFile.Close()
		close(done)
	}()

	return cmd.Process.Pid, done, nil
}

// ExecuteSync is like Execute, but synchronous.
func (ex *Executor) ExecuteSync(args ...string) ([]byte, error) {
	cmd := ex.generateCommand(args...)
	return cmd.Output()
}

// Plan runs 'terraform plan'.
func (ex *Executor) Plan() error {
	return ex.Execute("plan")
}

// Output gets output value from Terraform in JSON format and tries to unmarshal it
// to a given struct.
func (ex *Executor) Output(key string, s interface{}) error {
	o, err := ex.ExecuteSync("output", "-json", key)
	if err != nil {
		return fmt.Errorf("failed getting Terraform output: %w", err)
	}

	return json.Unmarshal(o, s)
}

// GenerateCommand prepares a Terraform command with the given arguments
// by setting up the command, configuration, working directory
// (so the files such as terraform.tfstate are stored at the right place) and
// extra environment variables. The current environment is fully inherited.
func (ex *Executor) generateCommand(args ...string) *exec.Cmd {
	cmd := exec.Command(ex.binaryPath, args...)
	// Copy environment because nil cannot be used to inherit if we add something in the next step.
	cmd.Env = os.Environ()
	for k, v := range ex.envVariables {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}
	cmd.Dir = ex.executionPath
	return cmd
}

// WorkingDirectory returns the directory in which Terraform runs, which can be
// useful for inspection or to retrieve any generated files.
func (ex *Executor) WorkingDirectory() string {
	return ex.executionPath
}

// Status returns the status of a given execution process.
//
// An error can be returned if the running processes could not be listed, or if
// the process failed, in which case the exit message is returned in an error
// of type ExecutionError.
//
// Note that if the identifier is invalid, the current implementation will
// return ExecutionStatusSuccess rather than ExecutionStatusUnknown.
func (ex *Executor) Status(id int) (ExecutionStatus, error) {
	isRunning, err := process.PidExists(int32(id))
	if err != nil {
		return ExecutionStatusUnknown, err
	}
	if isRunning {
		return ExecutionStatusRunning, nil
	}

	if failErr, err := ioutil.ReadFile(ex.failPath(id)); err == nil {
		return ExecutionStatusFailure, errors.New(string(failErr))
	}
	return ExecutionStatusSuccess, nil
}

// Ignore certain relative paths in the Terraform data dir. Paths must start at
// the top dir
var pathsToIgnore = map[string]struct{}{
	logsFolderName: {},
}

// tfBinatyPath searches for a Terraform binary on disk:
// - in the executing binary's folder,
// - in the current working directory,
// - in the PATH.
// The first to be found is the one returned.
func tfBinaryPath() (string, error) {
	// Depending on the platform, the expected binary name is different.
	binaryFileName := "terraform"
	if runtime.GOOS == "windows" {
		binaryFileName = "terraform.exe"
	}

	// Look into the executable's folder.
	if execFolderPath, err := osext.ExecutableFolder(); err == nil {
		path := filepath.Join(execFolderPath, binaryFileName)
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
			return path, nil
		}
	}

	// Look into cwd.
	if workingDirectory, err := os.Getwd(); err == nil {
		path := filepath.Join(workingDirectory, binaryFileName)
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
			return path, nil
		}
	}

	// If we still haven't found the executable, look for it
	// in the PATH.
	if path, err := exec.LookPath(binaryFileName); err == nil {
		return filepath.Abs(path)
	}

	return "", ErrBinaryNotFound
}

// failPath returns the path to the failure file of a given execution process.
func (ex *Executor) failPath(id int) string {
	failFileName := fmt.Sprintf("%d%s", id, failFileSuffix)
	return filepath.Join(ex.executionPath, logsFolderName, failFileName)
}

// logPath returns the path to the log file of a given execution process.
func (ex *Executor) logPath(id int) string {
	logFileName := fmt.Sprintf("%d%s", id, logsFileSuffix)
	return filepath.Join(ex.executionPath, logsFolderName, logFileName)
}

func (ex *Executor) checkVersion() error {
	vOutput, err := ex.ExecuteSync("--version")
	if err != nil {
		return fmt.Errorf("Error checking Terraform version: %w", err)
	}

	var vStr string
	n, err := fmt.Sscanf(string(vOutput), "Terraform v%s\n", &vStr)
	if err != nil {
		return fmt.Errorf("Error checking Terraform version: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("Parsing Terraform version failed")
	}

	v, err := version.NewVersion(vStr)
	if err != nil {
		return fmt.Errorf("Error checking Terraform version: %w", err)
	}

	constraints, err := version.NewConstraint(requiredVersion)
	if err != nil {
		return fmt.Errorf("Error checking Terraform version: %w", err)
	}

	if !constraints.Check(v) {
		return fmt.Errorf("Version '%s' of Terraform not supported. Needed %s", v, constraints)
	}

	return nil
}
