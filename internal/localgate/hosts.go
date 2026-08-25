package localgate

import "strings"

func containsHost(content string, domain string) bool {
	for _, line := range strings.Split(content, "\n") {
		for _, field := range strings.Fields(line) {
			if field == domain {
				return true
			}
		}
	}
	return false
}

func stringsWithoutManagedDomain(content string, domain string) string {
	lines := strings.Split(content, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, domain+" #localgate") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}
