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

func NewCopier[T any](getter func(v *T) any, setter func(o *T, v *T, nv any)) Copier[T] {
	return Copier[T]{
		memo:   map[*T]*T{},
		getter: getter,
		setter: setter,
	}
}

type Copier[T any] struct {
	memo   map[*T]*T
	setter func(o *T, v *T, nv any)
	getter func(v *T) any
}

func (c Copier[T]) Copy(v *T) *T {
	if v == nil {
		return nil
	}

	if copy, ok := c.memo[v]; ok {
		return copy
	}

	copy := new(T)
	c.memo[v] = copy

	var nv any
	switch vr := c.getter(v).(type) {
	case []T:
		a := make([]T, len(vr))
		for i, v := range vr {
			a[i] = *c.Copy(&v)
		}
		nv = a
	case map[string]T:
		m := make(map[string]T, len(vr))
		for k, v := range vr {
			m[k] = *c.Copy(&v)
		}
		nv = m
	default:
		nv = vr
	}

	c.setter(copy, v, nv)
	return copy
}
