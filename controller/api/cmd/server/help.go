// Package main contains the main work of the cloud controller api service.
package main

import "fmt"

func help() string {
	information := fmt.Sprintf("CYPHONIC cloud software controller v%s\n\nInformation available at https://github.com/Pluslab/cyphonic\nCopyright (C) @Pluslab, Lab. <%s>.\n",
		Version, CodeOwner)

	return information
}
