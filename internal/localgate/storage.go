package localgate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const HostFile = "/etc/hosts"

func DefaultConfigPath() string {
	home, err := userHome()
	if err != nil {
		return filepath.Join(".", ".localgate", "config.json")
	}
	return filepath.Join(home, ".localgate", "config.json")
}

func userHome() (string, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		account, err := user.Lookup(sudoUser)
		if err == nil {
			return account.HomeDir, nil
		}
	}
	return os.UserHomeDir()
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(Config), nil
	}
	if err != nil {
		return nil, fmt.Errorf("ler configuracao: %w", err)
	}

	config := make(Config)
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("ler configuracao: %w", err)
	}
	return config, nil
}

func saveConfig(path string, config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar configuracao: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("criar diretorio de configuracao: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("salvar configuracao: %w", err)
	}
	return nil
}

type HostsFile struct {
	path string
}

func NewHostsFile(path string) HostsFile {
	return HostsFile{path: path}
}

func (hosts HostsFile) Add(domain string) error {
	data, err := os.ReadFile(hosts.path)
	if err != nil {
		return fmt.Errorf("ler arquivo hosts: %w", err)
	}
	if containsHost(string(data), domain) {
		return nil
	}

	file, err := os.OpenFile(hosts.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("abrir arquivo hosts: %w", err)
	}
	defer file.Close()
	if _, err := fmt.Fprintf(file, "127.0.0.1 %s #localgate\n", domain); err != nil {
		return fmt.Errorf("atualizar arquivo hosts: %w", err)
	}
	return nil
}

func (hosts HostsFile) Remove(domain string) error {
	data, err := os.ReadFile(hosts.path)
	if err != nil {
		return fmt.Errorf("ler arquivo hosts: %w", err)
	}

	lines := stringsWithoutManagedDomain(string(data), domain)
	if err := os.WriteFile(hosts.path, []byte(lines), 0644); err != nil {
		return fmt.Errorf("atualizar arquivo hosts: %w", err)
	}
	return nil
}
