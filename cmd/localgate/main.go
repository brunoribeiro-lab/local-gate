package main

import (
	"os"

	"localgate/internal/localgate"
)

func main() {
	app := localgate.NewApp(
		os.Stdout,
		localgate.NewRouteService(
			localgate.DefaultConfigPath(),
			localgate.NewHostsFile(localgate.HostFile),
			localgate.NewNginx(),
		),
	)
	app.Run(os.Args[1:])
}
