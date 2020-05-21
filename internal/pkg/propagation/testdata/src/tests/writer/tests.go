// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package writer

import (
	"fmt"
	"strings"
)

type Source struct {
	Data string
	ID int
}

func TestPropagationOnVal(val Source) {
	buf := strings.Builder{} // want "this value becomes tainted"
	fmt.Fprintf(&buf, "%v", val)
}

func TestPropagationOnPtr(ptr *Source) {
	buf := strings.Builder{} // want "this value becomes tainted"
	fmt.Fprintf(&buf, "%v", ptr)
}
