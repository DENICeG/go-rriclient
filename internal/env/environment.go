package env

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"

	"github.com/sbreitf1/go-jcrypt"
)

const (
	envOrderFileName = "env-order"
)

// Environment represents an environment with address, user, password, and insecure flag.
type Environment struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"pass" jcrypt:"aes"`
	Insecure bool   `json:"insecure"`
}

func (e Environment) HasCredentials() bool {
	return len(e.User) > 0 && len(e.Password) > 0
}

type envOrder struct {
	Order []string `json:"order"`
	Fixed bool     `json:"fixed"`
}

// GetKeyHandler returns the encryption key for encryption or decryption.
type GetKeyHandler jcrypt.KeySource

// EnterEnvHandler prepares the environment defined by env.
type EnterEnvHandler func(env any) error

// GetEnvFileTitleHandler returns a human readable title for the given env file.
type GetEnvFileTitleHandler func(envName, envFile string) string

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

func GetEnvTitle(envName, envFile string) string {
	data, err := os.ReadFile(envFile)
	if err != nil {
		return envName
	}

	type envPreview struct {
		Address string `json:"address"`
		User    string `json:"user"`
	}
	var env envPreview
	if err := json.Unmarshal(data, &env); err != nil {
		return envName
	}

	var suffix string
	if len(env.User) > 0 {
		suffix = fmt.Sprintf(" (%s@%s)", env.User, env.Address)
	} else {
		suffix = fmt.Sprintf(" (%s)", env.Address)
	}
	return fmt.Sprintf("%s%s", envName, suffix)
}
