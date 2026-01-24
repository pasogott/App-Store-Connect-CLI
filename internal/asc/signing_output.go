package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func joinSigningList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}

func printSigningFetchResultTable(result *SigningFetchResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Bundle ID\tBundle ID Resource\tProfile Type\tProfile ID\tProfile File\tCertificate IDs\tCertificate Files\tCreated")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\n",
		result.BundleID,
		result.BundleIDResource,
		result.ProfileType,
		result.ProfileID,
		result.ProfileFile,
		joinSigningList(result.CertificateIDs),
		joinSigningList(result.CertificateFiles),
		result.Created,
	)
	return w.Flush()
}

func printSigningFetchResultMarkdown(result *SigningFetchResult) error {
	fmt.Fprintln(os.Stdout, "| Bundle ID | Bundle ID Resource | Profile Type | Profile ID | Profile File | Certificate IDs | Certificate Files | Created |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s | %s | %t |\n",
		escapeMarkdown(result.BundleID),
		escapeMarkdown(result.BundleIDResource),
		escapeMarkdown(result.ProfileType),
		escapeMarkdown(result.ProfileID),
		escapeMarkdown(result.ProfileFile),
		escapeMarkdown(joinSigningList(result.CertificateIDs)),
		escapeMarkdown(joinSigningList(result.CertificateFiles)),
		result.Created,
	)
	return nil
}
