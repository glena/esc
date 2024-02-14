// Copyright 2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewValue[T any](v T) PseudoValue {
	return PseudoValue{Value: v}
}

type PseudoValue struct {
	Value any
}

func TestCopy(t *testing.T) {
	a := NewValue(map[string]PseudoValue{
		"b": NewValue(map[string]PseudoValue{
			"c": NewValue("a.b"),
		}),
	})

	na := NewCopier(
		func(v *PseudoValue) any {
			return v.Value
		},
		func(o *PseudoValue, _ *PseudoValue, v any) {
			o.Value = v
		},
	).Copy(&a)

	newElement, ok := na.Value.(map[string]PseudoValue)
	assert.True(t, ok)
	newElement, ok = newElement["b"].Value.(map[string]PseudoValue)
	assert.True(t, ok)
	newElement["c"] = NewValue("1.2")
	newElement["d"] = NewValue("3.4")

	element, ok := a.Value.(map[string]PseudoValue)
	assert.True(t, ok)
	element, ok = element["b"].Value.(map[string]PseudoValue)
	assert.Equal(t, "a.b", element["c"].Value.(string))

	_, ok = element["d"]
	assert.False(t, ok)
}
