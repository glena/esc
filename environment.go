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

package esc

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pulumi/esc/schema"
)

const AnonymousEnvironmentName = "<yaml>"

type ExecContext struct {
	rootEvironment string
	values         map[string]Value
}

type copier struct {
	memo map[*Value]*Value
}

func newCopier() copier {
	return copier{memo: map[*Value]*Value{}}
}

func (c copier) copy(v *Value) *Value {
	if v == nil {
		return nil
	}

	if copy, ok := c.memo[v]; ok {
		return copy
	}

	copy := &Value{}
	c.memo[v] = copy

	var nv any
	switch vr := v.Value.(type) {
	case []*Value:
		a := make([]*Value, len(vr))
		for i, v := range vr {
			a[i] = c.copy(v)
		}
		nv = a
	case map[string]*Value:
		m := make(map[string]*Value, len(vr))
		for k, v := range vr {
			m[k] = c.copy(v)
		}
		nv = m
	default:
		nv = vr
	}

	*copy = Value{
		Value: nv,
	}
	return copy
}

func copyContext(context map[string]Value) map[string]Value {
	newContext := make(map[string]Value)
	for key, value := range context {
		copy := newCopier().copy(&value)
		newContext[key] = *copy
	}
	return newContext
}

func (ec *ExecContext) Copy() *ExecContext {
	return &ExecContext{
		values:         copyContext(ec.values),
		rootEvironment: ec.rootEvironment,
	}
}

func (ec *ExecContext) Values(envName string) map[string]Value {
	context := ec.values
	context["currentEnvironment"] = NewValue(map[string]Value{
		"name": NewValue(envName),
	})

	if ec.rootEvironment == AnonymousEnvironmentName || ec.rootEvironment == "" {
		ec.rootEvironment = envName
	}

	context["rootEnvironment"] = NewValue(map[string]Value{
		"name": NewValue(ec.rootEvironment),
	})

	return context
}

var ErrForbiddenContextkey = errors.New("forbidden context key")

func validateContextVariable(context map[string]Value, key string) error {
	if _, ok := context[key]; ok {
		return fmt.Errorf("%w: %q", ErrForbiddenContextkey, key)
	}
	return nil
}

func NewExecContext(values map[string]Value) (*ExecContext, error) {
	if err := validateContextVariable(values, "currentEnvironment"); err != nil {
		return nil, err
	}

	if err := validateContextVariable(values, "rootEnvironment"); err != nil {
		return nil, err
	}

	return &ExecContext{
		values: values,
	}, nil
}

// An Environment contains the result of evaluating an environment definition.
type Environment struct {
	// Exprs contains the AST for each expression in the environment definition.
	Exprs map[string]Expr `json:"exprs,omitempty"`

	// Properties contains the detailed values produced by the environment.
	Properties map[string]Value `json:"properties,omitempty"`

	// Schema contains the schema for Properties.
	Schema *schema.Schema `json:"schema,omitempty"`
}

// GetEnvironmentVariables returns any environment variables defined by the environment.
//
// Environment variables are any scalar values in the top-level `environmentVariables` property. Boolean and
// number values are converted to their string representation. The results remain Values in order to retain
// secret- and unknown-ness.
func (e *Environment) GetEnvironmentVariables() map[string]Value {
	obj, ok := e.Properties["environmentVariables"].Value.(map[string]Value)
	if !ok {
		return nil
	}

	var vars map[string]Value
	for k, v := range obj {
		switch v.Value.(type) {
		case nil, bool, json.Number, string:
			if vars == nil {
				vars = make(map[string]Value)
			}
			str := v.ToString(false)
			if v.Secret {
				vars[k] = NewSecret(str)
			} else {
				vars[k] = NewValue(str)
			}
		}
	}
	return vars
}

// GetTemporaryFiles returns any temporary files defined by the environment.
//
// Temporary files are any string values in the top-level `files` property. The key for each file is the name
// of the environment variable that should hold the actual path to the temporary file once it is written.
func (e *Environment) GetTemporaryFiles() map[string]Value {
	obj, ok := e.Properties["files"].Value.(map[string]Value)
	if !ok {
		return nil
	}

	var files map[string]Value
	for k, v := range obj {
		switch v.Value.(type) {
		case nil, bool, json.Number, string:
			if files == nil {
				files = make(map[string]Value)
			}
			str := v.ToString(false)
			if v.Secret {
				files[k] = NewSecret(str)
			} else {
				files[k] = NewValue(str)
			}
		}
	}
	return files
}
