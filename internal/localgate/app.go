package localgate

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

type Routes interface {
	Add(string, int) error
	Remove(string) error
	List() (Config, error)
	Reload() error
}

type App struct {
	out     io.Writer
	routes  Routes
	isAdmin func() bool
}

func NewApp(out io.Writer, routes Routes) App {
	return NewAppWithPrivilege(out, routes, isAdministrator)
}

func NewAppWithPrivilege(out io.Writer, routes Routes, isAdmin func() bool) App {
	return App{out: out, routes: routes, isAdmin: isAdmin}
}

func isAdministrator() bool {
	return os.Geteuid() == 0
}

func (app App) Run(args []string) {
	if len(args) == 0 {
		app.printHelp()
		return
	}

	switch args[0] {
	case "help", "--help", "-h":
		app.printHelp()
	case "add":
		app.add(args)
	case "remove", "del":
		app.remove(args)
	case "list":
		app.list()
	case "start":
		app.reload()
	default:
		fmt.Fprintf(app.out, "Comando desconhecido: %s\n\n", args[0])
		app.printHelp()
	}
}

func (app App) add(args []string) {
	if len(args) != 3 {
		fmt.Fprintln(app.out, "Uso: localgate add <nome> <porta>")
		return
	}
	if !app.requireAdministrator() {
		return
	}

	port, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintln(app.out, "Porta invalida.")
		return
	}
	if err := app.routes.Add(args[1], port); err != nil {
		fmt.Fprintf(app.out, "Erro ao adicionar dominio: %v\n", err)
		return
	}
	fmt.Fprintf(app.out, "Adicionado: http://%s -> localhost:%d\n", DomainForName(args[1]), port)
}

func (app App) remove(args []string) {
	if len(args) != 2 {
		fmt.Fprintln(app.out, "Uso: localgate remove <nome>")
		return
	}
	if !app.requireAdministrator() {
		return
	}
	if err := app.routes.Remove(args[1]); err != nil {
		fmt.Fprintf(app.out, "Erro ao remover dominio: %v\n", err)
		return
	}
	fmt.Fprintf(app.out, "Removido: %s\n", DomainForName(args[1]))
}

func (app App) list() {
	config, err := app.routes.List()
	if err != nil {
		fmt.Fprintf(app.out, "Erro ao listar dominios: %v\n", err)
		return
	}

	names := make([]string, 0, len(config))
	for name := range config {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintln(app.out, "Nenhum dominio configurado.")
		return
	}
	for _, name := range names {
		fmt.Fprintf(app.out, "http://%s -> localhost:%d\n", DomainForName(name), config[name])
	}
}

func (app App) reload() {
	if !app.requireAdministrator() {
		return
	}
	if err := app.routes.Reload(); err != nil {
		fmt.Fprintf(app.out, "Erro ao recarregar Nginx: %v\n", err)
		return
	}
	fmt.Fprintln(app.out, "Nginx validado e recarregado.")
}

func (app App) requireAdministrator() bool {
	if app.isAdmin() {
		return true
	}
	fmt.Fprintln(app.out, "Este comando requer privilegios de administrador. Execute com sudo localgate.")
	return false
}

func (app App) printHelp() {
	help := "Localgate - Proxy Reverso Local para Desenvolvimento\n\n" +
		"Uso:\n" +
		"  localgate <comando> [argumentos]\n\n" +
		"Comandos:\n" +
		"  add <nome> <porta>    Adiciona um dominio apontando para a porta local.\n" +
		"  remove, del <nome>    Remove um dominio existente.\n" +
		"  list                  Lista todos os dominios configurados e suas portas.\n" +
		"  start                 Valida e recarrega as configuracoes do Nginx.\n" +
		"  help                  Mostra esta mensagem de ajuda.\n\n" +
		"Os comandos add, remove e start requerem sudo."
	fmt.Fprintln(app.out, help)
}
