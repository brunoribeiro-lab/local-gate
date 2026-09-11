package localgate

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"
)

type fakeRoutes struct {
	calls  []string
	config Config
	err    error
}

func (routes *fakeRoutes) Add(name string, port int) error {
	routes.calls = append(routes.calls, "add:"+name+":"+strconv.Itoa(port))
	return routes.err
}

func (routes *fakeRoutes) Remove(name string) error {
	routes.calls = append(routes.calls, "remove:"+name)
	return routes.err
}

func (routes *fakeRoutes) List() (Config, error) {
	routes.calls = append(routes.calls, "list")
	return routes.config, routes.err
}

func (routes *fakeRoutes) Reload() error {
	routes.calls = append(routes.calls, "reload")
	return routes.err
}

func newTestApp(out *bytes.Buffer, routes Routes) App {
	return NewAppWithPrivilege(out, routes, func() bool { return true })
}

func TestAppCommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCall string
		wantText string
	}{
		{"help", []string{"help"}, "", "Uso:"},
		{"empty", nil, "", "Uso:"},
		{"add", []string{"add", "api", "3000"}, "add:api:3000", "Adicionado: http://api.local"},
		{"add full domain", []string{"add", "admin-api.monitorstack.online", "9092"}, "add:admin-api.monitorstack.online:9092", "Adicionado: http://admin-api.monitorstack.online -> localhost:9092"},
		{"add arguments", []string{"add", "api"}, "", "Uso: localgate add"},
		{"add invalid port", []string{"add", "api", "x"}, "", "Porta invalida."},
		{"remove", []string{"remove", "api"}, "remove:api", "Removido: api.local"},
		{"remove full domain", []string{"remove", "app.com.br"}, "remove:app.com.br", "Removido: app.com.br"},
		{"del alias", []string{"del", "api"}, "remove:api", "Removido: api.local"},
		{"remove arguments", []string{"remove"}, "", "Uso: localgate remove"},
		{"list", []string{"list"}, "list", "http://api.local -> localhost:3000"},
		{"start", []string{"start"}, "reload", "Nginx validado e recarregado."},
		{"unknown", []string{"other"}, "", "Comando desconhecido: other"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routes := &fakeRoutes{config: Config{"api": 3000}}
			output := &bytes.Buffer{}
			newTestApp(output, routes).Run(test.args)
			if !strings.Contains(output.String(), test.wantText) {
				t.Fatalf("output = %q, want text %q", output.String(), test.wantText)
			}
			if test.wantCall == "" && len(routes.calls) != 0 {
				t.Fatalf("calls = %v, want no calls", routes.calls)
			}
			if test.wantCall != "" && (len(routes.calls) != 1 || routes.calls[0] != test.wantCall) {
				t.Fatalf("calls = %v, want %q", routes.calls, test.wantCall)
			}
		})
	}
}

func TestAppReportsCommandErrors(t *testing.T) {
	routes := &fakeRoutes{err: errors.New("failure")}
	output := &bytes.Buffer{}
	newTestApp(output, routes).Run([]string{"start"})
	if !strings.Contains(output.String(), "Erro ao recarregar Nginx: failure") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestAppRequiresAdministrator(t *testing.T) {
	output := &bytes.Buffer{}
	routes := &fakeRoutes{}
	NewAppWithPrivilege(output, routes, func() bool { return false }).Run([]string{"add", "api", "3000"})
	if !strings.Contains(output.String(), "requer privilegios") {
		t.Fatalf("output = %q", output.String())
	}
	if len(routes.calls) != 0 {
		t.Fatalf("calls = %v, want no calls", routes.calls)
	}
}
