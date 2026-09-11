package localgate

import (
	"fmt"
	"strings"
)

const Suffix = ".local"

type Config map[string]int

type HostManager interface {
	Add(string) error
	Remove(string) error
}

type NginxManager interface {
	Ensure() error
	Add(string, int) error
	Remove(string) error
	Reload() error
}

type RouteService struct {
	configPath string
	hosts      HostManager
	nginx      NginxManager
}

func NewRouteService(configPath string, hosts HostManager, nginx NginxManager) RouteService {
	return RouteService{configPath: configPath, hosts: hosts, nginx: nginx}
}

func (service RouteService) Add(name string, port int) error {
	if err := validateRoute(name, port); err != nil {
		return err
	}
	if err := service.nginx.Ensure(); err != nil {
		return err
	}

	domain := DomainForName(name)
	if err := service.nginx.Add(domain, port); err != nil {
		return err
	}
	if err := service.nginx.Reload(); err != nil {
		return err
	}
	if err := service.hosts.Add(domain); err != nil {
		return err
	}

	config, err := loadConfig(service.configPath)
	if err != nil {
		return err
	}
	config[name] = port
	return saveConfig(service.configPath, config)
}

func (service RouteService) Remove(name string) error {
	if err := validateName(name); err != nil {
		return err
	}

	config, err := loadConfig(service.configPath)
	if err != nil {
		return err
	}
	if _, exists := config[name]; !exists {
		return fmt.Errorf("dominio %s nao foi configurado pelo Localgate", DomainForName(name))
	}
	if err := service.nginx.Ensure(); err != nil {
		return err
	}

	domain := DomainForName(name)
	if err := service.nginx.Remove(domain); err != nil {
		return err
	}
	if err := service.nginx.Reload(); err != nil {
		return err
	}
	if err := service.hosts.Remove(domain); err != nil {
		return err
	}

	delete(config, name)
	return saveConfig(service.configPath, config)
}

func (service RouteService) List() (Config, error) {
	return loadConfig(service.configPath)
}

func (service RouteService) Reload() error {
	if err := service.nginx.Ensure(); err != nil {
		return err
	}
	return service.nginx.Reload()
}

func DomainForName(name string) string {
	if strings.Contains(name, ".") {
		return name
	}
	return name + Suffix
}

func validateRoute(name string, port int) error {
	if err := validateName(name); err != nil {
		return err
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("porta invalida: %d", port)
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("nome do dominio nao pode ser vazio")
	}
	for _, label := range strings.Split(name, ".") {
		if !validLabel(label) {
			return fmt.Errorf("nome de dominio invalido: %s", name)
		}
	}
	return nil
}

func validLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, character := range label {
		if character != '-' && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}
