package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// SandboxTesterDeleteResult represents CLI output for sandbox tester deletions.
type SandboxTesterDeleteResult struct {
	ID      string `json:"id"`
	Email   string `json:"email,omitempty"`
	Deleted bool   `json:"deleted"`
}

func formatSandboxTesterName(attr SandboxTesterAttributes) string {
	return compactWhitespace(strings.TrimSpace(attr.FirstName + " " + attr.LastName))
}

func printSandboxTestersTable(resp *SandboxTestersResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEmail\tName\tTerritory")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.ID,
			sandboxTesterEmail(item.Attributes),
			formatSandboxTesterName(item.Attributes),
			sandboxTesterTerritory(item.Attributes),
		)
	}
	return w.Flush()
}

func printSandboxTestersMarkdown(resp *SandboxTestersResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Email | Name | Territory |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(sandboxTesterEmail(item.Attributes)),
			escapeMarkdown(formatSandboxTesterName(item.Attributes)),
			escapeMarkdown(sandboxTesterTerritory(item.Attributes)),
		)
	}
	return nil
}

func printSandboxTesterDeleteResultTable(result *SandboxTesterDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEmail\tDeleted")
	fmt.Fprintf(w, "%s\t%s\t%t\n",
		result.ID,
		result.Email,
		result.Deleted,
	)
	return w.Flush()
}

func printSandboxTesterDeleteResultMarkdown(result *SandboxTesterDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Email | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %t |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.Email),
		result.Deleted,
	)
	return nil
}
