// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package v2

import (
	"crypto/sha256"
	"fmt"
	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/index"
	"gopkg.in/yaml.v3"
	"sort"
	"strings"
)

// SecurityRequirement is a low-level representation of a Swagger / OpenAPI 2 SecurityRequirement object.
//
// SecurityRequirement lists the required security schemes to execute this operation. The object can have multiple
// security schemes declared in it which are all required (that is, there is a logical AND between the schemes).
//
// The name used for each property MUST correspond to a security scheme declared in the Security Definitions
//  - https://swagger.io/specification/v2/#securityDefinitionsObject
type SecurityRequirement struct {
	Values low.ValueReference[map[low.KeyReference[string]]low.ValueReference[[]low.ValueReference[string]]]
}

// Build will extract security requirements from the node (the structure is odd, to be honest)
func (s *SecurityRequirement) Build(root *yaml.Node, _ *index.SpecIndex) error {
	var labelNode *yaml.Node
	valueMap := make(map[low.KeyReference[string]]low.ValueReference[[]low.ValueReference[string]])
	var arr []low.ValueReference[string]
	for i := range root.Content {
		if i%2 == 0 {
			labelNode = root.Content[i]
			arr = []low.ValueReference[string]{} // reset roles.
			continue
		}
		for j := range root.Content[i].Content {
			arr = append(arr, low.ValueReference[string]{
				Value:     root.Content[i].Content[j].Value,
				ValueNode: root.Content[i].Content[j],
			})
		}
		valueMap[low.KeyReference[string]{
			Value:   labelNode.Value,
			KeyNode: labelNode,
		}] = low.ValueReference[[]low.ValueReference[string]]{
			Value:     arr,
			ValueNode: root.Content[i],
		}
	}
	s.Values = low.ValueReference[map[low.KeyReference[string]]low.ValueReference[[]low.ValueReference[string]]]{
		Value:     valueMap,
		ValueNode: root,
	}
	return nil
}

// Hash will return a consistent SHA256 Hash of the SecurityRequirement object
func (s *SecurityRequirement) Hash() [32]byte {
	var f []string
	values := make(map[string][]string, len(s.Values.Value))
	var valKeys []string
	for k := range s.Values.Value {
		var vals []string
		for y := range s.Values.Value[k].Value {
			vals = append(vals, s.Values.Value[k].Value[y].Value)
			// lol, I know. -------^^^^^ <- this is the actual value.
		}
		sort.Strings(vals)
		valKeys = append(valKeys, k.Value)
		if len(vals) > 0 {
			values[k.Value] = vals
		}
	}
	sort.Strings(valKeys)
	for val := range valKeys {
		f = append(f, fmt.Sprintf("%s-%s", valKeys[val], strings.Join(values[valKeys[val]], "|")))
	}
	return sha256.Sum256([]byte(strings.Join(f, "|")))
}
