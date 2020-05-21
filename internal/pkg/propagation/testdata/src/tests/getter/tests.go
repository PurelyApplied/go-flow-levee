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

package getter

type Source struct {
	Data string
	ID   int
}

func (s Source) ValGetData() string {
	return s.Data
}

func (s Source) ValGetID() int {
	return s.ID
}

type SourcePtr struct {
	Data string
	ID   int
}

func (s *SourcePtr) PtrGetData() string {
	return s.Data
}

func (s *SourcePtr) PtrGetID() int {
	return s.ID
}

func noop(...interface{}) {}

func TestGetters(val Source, ptr Source, sp SourcePtr, pp *SourcePtr) {
	t1 := val.ValGetData() // want "this value becomes tainted"
	t2 := ptr.ValGetData() // want "this value becomes tainted"
	t3 := sp.PtrGetData()  // TODO want "this value becomes tainted"
	t4 := pp.PtrGetData()  // TODO want "this value becomes tainted"

	ok1 := val.ValGetID()
	ok2 := ptr.ValGetID()
	ok3 := sp.PtrGetID()
	ok4 := pp.PtrGetID()

	noop(t1, t2, t3, t4, ok1, ok2, ok3, ok4)
}
