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

package prompt

import (
	"io"

	"github.com/manifoldco/promptui"
)

func NewPrompt(label string, validate promptui.ValidateFunc, in io.Reader) *promptui.Prompt {
	template := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	if validate == nil {
		template = &promptui.PromptTemplates{
			Prompt:  "{{ . }} ",
			Valid:   "{{ . | red }} ",
			Invalid: "{{ . | red }} ",
			Success: "{{ . | bold }} ",
		}
	}
	p := promptui.Prompt{
		Label:     label,
		Stdin:     io.NopCloser(in),
		Templates: template,
		Validate:  validate,
	}
	return &p
}
