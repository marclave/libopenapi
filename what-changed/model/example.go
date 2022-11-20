// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package model

import (
	"fmt"
	"github.com/pb33f/libopenapi/datamodel/low/base"
	v3 "github.com/pb33f/libopenapi/datamodel/low/v3"
	"github.com/pb33f/libopenapi/utils"
	"sort"
)

// ExampleChanges represent changes to an Example object, part of an OpenAPI specification.
type ExampleChanges struct {
	PropertyChanges
	ExtensionChanges *ExtensionChanges `json:"extensions,omitempty" yaml:"extensions,omitempty"`
}

// TotalChanges returns the total number of changes made to Example
func (e *ExampleChanges) TotalChanges() int {
	l := e.PropertyChanges.TotalChanges()
	if e.ExtensionChanges != nil {
		l += e.ExtensionChanges.PropertyChanges.TotalChanges()
	}
	return l
}

// TotalBreakingChanges returns the total number of breaking changes made to Example
func (e *ExampleChanges) TotalBreakingChanges() int {
	l := e.PropertyChanges.TotalBreakingChanges()
	return l
}

// TotalChanges

func CompareExamples(l, r *base.Example) *ExampleChanges {

	ec := new(ExampleChanges)
	var changes []*Change
	var props []*PropertyCheck

	// Summary
	props = append(props, &PropertyCheck{
		LeftNode:  l.Summary.ValueNode,
		RightNode: r.Summary.ValueNode,
		Label:     v3.SummaryLabel,
		Changes:   &changes,
		Breaking:  false,
		Original:  l,
		New:       r,
	})

	// Description
	props = append(props, &PropertyCheck{
		LeftNode:  l.Description.ValueNode,
		RightNode: r.Description.ValueNode,
		Label:     v3.DescriptionLabel,
		Changes:   &changes,
		Breaking:  false,
		Original:  l,
		New:       r,
	})

	// Value
	if utils.IsNodeMap(l.Value.ValueNode) && utils.IsNodeMap(r.Value.ValueNode) {
		lKeys := make([]string, len(l.Value.ValueNode.Content)/2)
		rKeys := make([]string, len(r.Value.ValueNode.Content)/2)
		z := 0
		for k := range l.Value.ValueNode.Content {
			if k%2 == 0 {
				lKeys[z] = fmt.Sprintf("%v-%v", l.Value.ValueNode.Content[k].Value, l.Value.ValueNode.Content[k+1].Value)
				z++
			} else {
				continue
			}
		}
		z = 0
		for k := range r.Value.ValueNode.Content {
			if k%2 == 0 {
				rKeys[z] = fmt.Sprintf("%v-%v", r.Value.ValueNode.Content[k].Value, r.Value.ValueNode.Content[k+1].Value)
				z++
			} else {
				continue
			}
		}
		sort.Strings(lKeys)
		sort.Strings(rKeys)
		if (len(lKeys) > len(rKeys)) || (len(rKeys) > len(lKeys)) {
			CreateChange(&changes, Modified, v3.ValueLabel,
				l.Value.GetValueNode(), r.Value.GetValueNode(), false, l.Value.GetValue(), r.Value.GetValue())
		}
		for k := range lKeys {
			if lKeys[k] != rKeys[k] {
				CreateChange(&changes, Modified, v3.ValueLabel,
					l.Value.GetValueNode(), r.Value.GetValueNode(), false, l.Value.GetValue(), r.Value.GetValue())
			}
		}
		for k := range rKeys {
			if k >= len(lKeys) {
				CreateChange(&changes, Modified, v3.ValueLabel,
					nil, r.Value.GetValueNode(), false, nil, r.Value.GetValue())
			}
		}
	} else {
		props = append(props, &PropertyCheck{
			LeftNode:  l.Value.ValueNode,
			RightNode: r.Value.ValueNode,
			Label:     v3.ValueLabel,
			Changes:   &changes,
			Breaking:  false,
			Original:  l,
			New:       r,
		})
	}
	// ExternalValue
	props = append(props, &PropertyCheck{
		LeftNode:  l.ExternalValue.ValueNode,
		RightNode: r.ExternalValue.ValueNode,
		Label:     v3.ExternalValue,
		Changes:   &changes,
		Breaking:  false,
		Original:  l,
		New:       r,
	})

	// check properties
	CheckProperties(props)

	// check extensions
	ec.ExtensionChanges = CheckExtensions(l, r)
	ec.Changes = changes
	if ec.TotalChanges() <= 0 {
		return nil
	}
	return ec
}
