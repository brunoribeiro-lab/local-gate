# Localgate

Localgate e uma CLI Go para criar dominios `.local` para aplicacoes de desenvolvimento. Cada rota gera um virtual host Nginx que encaminha o trafego para uma porta local.

## Como Funciona

![Diagrama de Arquitetura](docs/diagram.jpeg)

O projeto atua em três frentes principais:
1. **Resolução DNS Local:** Modifica o `/etc/hosts` para que domínios `.local` apontem para `127.0.0.1`.
2. **Proxy Reverso:** Intercepta a requisição na porta `80` (HTTP).
3. **Roteamento:** Lê o cabeçalho HTTP da requisição e repassa o tráfego para a porta correta da sua aplicação (ex: `3000`).

## Funcionalidades

- Cria entradas em `/etc/hosts` e virtual hosts Nginx.
- Valida a configuracao com `nginx -t` antes de recarregar o servico.
- Aceita nomes com subdominios, como `api.site`.
- Mantem as rotas do Localgate em `~/.localgate/config.json`.

## Instalacao

### Pré-requisitos
- [Go instalado](https://go.dev/doc/install) (versão 1.18 ou superior).
- Nginx instalado e com os diretórios `/etc/nginx/sites-available` e `/etc/nginx/sites-enabled` configurados.

### Passo a passo

1. Clone o repositório ou crie o diretório do projeto:
   ```bash
   mkdir localgate
   cd localgate
   ```

2. Compile e execute os testes:
   ```bash
   go test ./...
   go build -o localgate ./cmd/localgate
   ```

3. Mova o binário gerado para a pasta do sistema para utilizá-lo de qualquer lugar:
   ```bash
   sudo mv localgate /usr/local/bin/
   ```

---

## Uso

O fluxo de uso básico envolve adicionar as rotas. O Localgate cria e habilita um virtual host do Nginx para cada domínio.

### Adicionar dominios
Informe o nome sem `.local` e a porta onde a aplicacao sera executada.
```bash
sudo localgate add front 3000
sudo localgate add api 4000
```
Isso cria `front.local` apontando para `localhost:3000`.

Agora você já pode abrir no seu navegador: **`http://front.local`**.

### Outros comandos

**Listar dominios configurados:**
```bash
localgate list
```

**Remover um dominio:**
```bash
sudo localgate remove front
# Alias curto:
sudo localgate del front
```

**Exibir ajuda:**
```bash
localgate help
```

---

## Arquitetura

O executavel fica em `cmd/localgate/main.go`. A implementacao fica em `internal/localgate`: `app.go` trata argumentos e saida; `routes.go` coordena o cadastro; `storage.go` administra configuracao e hosts; e `nginx.go` administra os virtual hosts. Essa separacao permite testar todos os comandos sem escrever em `/etc` nem iniciar o Nginx.

## Sobre go.mod

`go.mod` e o arquivo padrao do Go para declarar um modulo: seu nome e fixo porque o comando `go` o procura automaticamente. A linha `module github.com/fr4nk/localgate` informa o caminho usado para importar este projeto. Quando o repositorio for criado no GitHub, mantenha esse valor se a URL for `github.com/fr4nk/localgate`; caso o usuario ou repositorio seja diferente, atualize-o para a URL real.