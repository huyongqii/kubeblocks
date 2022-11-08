/*
Copyright ApeCloud Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/apecloud/kubeblocks/internal/dbctl/cmd"
)

func main() {
	rootPath := "./docs/cli"
	if len(os.Args) > 1 {
		rootPath = os.Args[1]
	}

	dbctl := cmd.NewDbctlCmd()
	dbctl.Long = fmt.Sprintf("```\n%s\n```", dbctl.Long)
	err := doc.GenMarkdownTree(dbctl, rootPath)
	if err != nil {
		log.Fatal(err)
	}
}