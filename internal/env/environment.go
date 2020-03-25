package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sbreitf1/go-console"
	"github.com/sbreitf1/go-jcrypt"
)

//TODO save env ordering (latest used)

// GetKeyHandler returns the encryption key for encryption or decryption.
type GetKeyHandler jcrypt.KeySource

// EnterEnvHandler prepares the environment defined by env.
type EnterEnvHandler func(envName string, env interface{}) error

// GetEnvFileTitleHandler returns a humand readable title for the given env file.
type GetEnvFileTitleHandler func(envName, envFile string) string

// Reader represents a reader object for environments.
type Reader struct {
	dir             string
	KeySource       GetKeyHandler
	EnterEnvHandler EnterEnvHandler
	GetEnvFileTitle GetEnvFileTitleHandler
}

// NewReader returns a new environment reader using the given key source and enter environment handler.
func NewReader(homeDirName string) (*Reader, error) {
	dir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	return &Reader{dir: filepath.Join(dir, homeDirName)}, nil
}

func (e *Reader) getEnvFilePath(envName string) string {
	//TODO conditional escaping
	return filepath.Join(e.dir, envName+".json")
}

// ReadEnvironment reads an existing environment.
func (e *Reader) ReadEnvironment(envName string, env interface{}) error {
	return e.createOrReadEnvironment(envName, env, nil)
}

// CreateOrReadEnvironment reads an existing environment with the given name or calls the enter environment reader.
func (e *Reader) CreateOrReadEnvironment(envName string, env interface{}) error {
	return e.createOrReadEnvironment(envName, env, e.EnterEnvHandler)
}

func (e *Reader) createOrReadEnvironment(envName string, env interface{}, enterEnvHandler EnterEnvHandler) error {
	file := e.getEnvFilePath(envName)
	exists, err := isFile(file)
	if err != nil {
		return err
	}

	keySource := jcrypt.KeySource(e.KeySource)
	if keySource == nil {
		keySource = func() ([]byte, error) { return []byte{}, nil }
	}

	if !exists {
		if enterEnvHandler != nil {
			console.Printlnf("Environment %q does not exist yet, pleaser enter below:", envName)
			err := enterEnvHandler(envName, env)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(e.dir, os.ModePerm); err != nil {
				console.Printlnf("WARNING: failed to save environment: %s", err.Error())
			} else {
				if err := jcrypt.MarshalToFile(file, env, &jcrypt.Options{
					GetKeyHandler: keySource,
				}); err != nil {
					console.Printlnf("WARNING: failed to save environment: %s", err.Error())
				}
			}

			return nil
		}
		return fmt.Errorf("environment %q not found", envName)
	}

	if err := jcrypt.UnmarshalFromFile(file, env, &jcrypt.Options{
		GetKeyHandler: keySource,
	}); err != nil {
		return err
	}
	return nil
}

// SelectEnvironment displays all configured environments and prompts the user.
func (e *Reader) SelectEnvironment(env interface{}) error {
	envFiles, err := e.GetEnvironmentFiles()
	if err != nil {
		return err
	}
	if len(envFiles) == 0 {
		return fmt.Errorf("no environments specified")
	}

	envTitles := make([]string, len(envFiles))
	for i, fi := range envFiles {
		name := fi.Name()
		if strings.HasSuffix(name, ".json") {
			name = name[:len(name)-5]
		}

		if e.GetEnvFileTitle != nil {
			envTitles[i] = e.GetEnvFileTitle(name, filepath.Join(e.dir, fi.Name()))
		} else {
			envTitles[i] = name
		}
	}
	ui := promptui.Select{Label: "Select environment", Items: envTitles}
	index, _, err := ui.Run()
	if err != nil {
		return err
	}

	fileName := envFiles[index].Name()
	return e.createOrReadEnvironment(fileName[:len(fileName)-5], env, nil)
}

// GetEnvironmentFiles returns all files that contain environments.
func (e *Reader) GetEnvironmentFiles() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(e.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []os.FileInfo{}, nil
		}
		return nil, err
	}

	envFiles := make([]os.FileInfo, 0)
	for _, fi := range files {
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".json") {
			envFiles = append(envFiles, fi)
		}
	}
	return envFiles, nil
}

func getConfigDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func isFile(file string) (bool, error) {
	fi, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !fi.IsDir(), nil
}
