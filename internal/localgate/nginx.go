package localgate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	NginxSitesAvail   = "/etc/nginx/sites-available"
	NginxSitesEnabled = "/etc/nginx/sites-enabled"
)

type commandRunner func(string, ...string) ([]byte, error)

type Nginx struct {
	available string
	enabled   string
	lookPath  func(string) (string, error)
	run       commandRunner
}

func NewNginx() Nginx {
	return Nginx{
		available: NginxSitesAvail,
		enabled:   NginxSitesEnabled,
		lookPath:  exec.LookPath,
		run:       runCommand,
	}
}

func runCommand(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func (nginx Nginx) Ensure() error {
	if _, err := nginx.lookPath("nginx"); err != nil {
		return fmt.Errorf("Nginx nao esta instalado ou nao esta no PATH")
	}
	for _, directory := range []string{nginx.available, nginx.enabled} {
		info, err := os.Stat(directory)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("diretorio do Nginx nao encontrado: %s", directory)
		}
	}
	return nil
}

func (nginx Nginx) Add(domain string, port int) error {
	sitePath := filepath.Join(nginx.available, domain)
	if err := os.WriteFile(sitePath, []byte(nginxConfig(domain, port)), 0644); err != nil {
		return fmt.Errorf("criar configuracao do Nginx: %w", err)
	}

	enabledPath := filepath.Join(nginx.enabled, domain)
	if err := os.Symlink(sitePath, enabledPath); err != nil && !os.IsExist(err) {
		return fmt.Errorf("habilitar configuracao do Nginx: %w", err)
	}
	return nil
}

func (nginx Nginx) Remove(domain string) error {
	paths := []string{
		filepath.Join(nginx.enabled, domain),
		filepath.Join(nginx.available, domain),
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remover configuracao do Nginx: %w", err)
		}
	}
	return nil
}

func (nginx Nginx) Reload() error {
	if output, err := nginx.run("nginx", "-t"); err != nil {
		return fmt.Errorf("configuracao invalida: %s", strings.TrimSpace(string(output)))
	}
	if output, err := nginx.run("systemctl", "reload", "nginx"); err != nil {
		return fmt.Errorf("falha ao recarregar Nginx: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func nginxConfig(domain string, port int) string {
	return fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://127.0.0.1:%d;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`, domain, port)
}
