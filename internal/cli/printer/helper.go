/*
Copyright ApeCloud, Inc.

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

package printer

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// BoldYellow returns a string formatted with yellow and bold.
func BoldYellow(msg interface{}) string {
	return color.New(color.FgYellow).Add(color.Bold).Sprint(msg)
}

// BoldRed returns a string formatted with red and bold.
func BoldRed(msg interface{}) string {
	return color.New(color.FgRed).Add(color.Bold).Sprint(msg)
}

// BoldGreen returns a string formatted with green and bold.
func BoldGreen(msg interface{}) string {
	return color.New(color.FgGreen, color.Bold).Sprint(msg)
}

func Warning(out io.Writer, format string, i ...interface{}) {
	fmt.Fprintf(out, "%s %s", BoldYellow("Warning:"), fmt.Sprintf(format, i...))
}

func Alert(out io.Writer, format string, i ...interface{}) {
	fmt.Fprintf(out, "%s %s", BoldRed("Alert:"), fmt.Sprintf(format, i...))
}
