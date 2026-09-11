package localgate

import (
	"path/filepath"
	"testing"
)

type fakeHosts struct {
	added   []string
	removed []string
}

func (hosts *fakeHosts) Add(domain string) error {
	hosts.added = append(hosts.added, domain)
	return nil
}

func (hosts *fakeHosts) Remove(domain string) error {
	hosts.removed = append(hosts.removed, domain)
	return nil
}

type fakeNginx struct {
	added   []string
	removed []string
	reloads int
	ensures int
}

func (nginx *fakeNginx) Ensure() error {
	nginx.ensures++
	return nil
}

func (nginx *fakeNginx) Add(domain string, port int) error {
	nginx.added = append(nginx.added, domain)
	return nil
}

func (nginx *fakeNginx) Remove(domain string) error {
	nginx.removed = append(nginx.removed, domain)
	return nil
}

func (nginx *fakeNginx) Reload() error {
	nginx.reloads++
	return nil
}

func TestRouteServiceAddsAndRemovesRoute(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	hosts := &fakeHosts{}
	nginx := &fakeNginx{}
	service := NewRouteService(configPath, hosts, nginx)

	if err := service.Add("api", 3000); err != nil {
		t.Fatal(err)
	}
	config, err := service.List()
	if err != nil || config["api"] != 3000 {
		t.Fatalf("config = %v, err = %v", config, err)
	}
	if len(hosts.added) != 1 || hosts.added[0] != "api.local" {
		t.Fatalf("added hosts = %v", hosts.added)
	}

	if err := service.Remove("api"); err != nil {
		t.Fatal(err)
	}
	config, err = service.List()
	if err != nil || len(config) != 0 {
		t.Fatalf("config = %v, err = %v", config, err)
	}
	if len(hosts.removed) != 1 || hosts.removed[0] != "api.local" {
		t.Fatalf("removed hosts = %v", hosts.removed)
	}
}

func TestRouteServiceKeepsFullDomain(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	hosts := &fakeHosts{}
	nginx := &fakeNginx{}
	service := NewRouteService(configPath, hosts, nginx)

	name := "admin-api.monitorstack.online"
	if err := service.Add(name, 9092); err != nil {
		t.Fatal(err)
	}
	if len(hosts.added) != 1 || hosts.added[0] != name {
		t.Fatalf("added hosts = %v", hosts.added)
	}
	if len(nginx.added) != 1 || nginx.added[0] != name {
		t.Fatalf("added nginx = %v", nginx.added)
	}

	if err := service.Remove(name); err != nil {
		t.Fatal(err)
	}
	if len(hosts.removed) != 1 || hosts.removed[0] != name {
		t.Fatalf("removed hosts = %v", hosts.removed)
	}
	if len(nginx.removed) != 1 || nginx.removed[0] != name {
		t.Fatalf("removed nginx = %v", nginx.removed)
	}
}

func TestRouteServiceRejectsInvalidRoutes(t *testing.T) {
	service := NewRouteService(filepath.Join(t.TempDir(), "config.json"), &fakeHosts{}, &fakeNginx{})
	for _, test := range []struct {
		name string
		port int
	}{
		{"bad/name", 3000},
		{"api", 0},
		{"api", 65536},
	} {
		if err := service.Add(test.name, test.port); err == nil {
			t.Fatalf("Add(%q, %d) succeeded", test.name, test.port)
		}
	}
}
