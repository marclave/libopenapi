// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package model

import (
	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/datamodel/low/base"
	"github.com/pb33f/libopenapi/datamodel/low/v2"
	"github.com/pb33f/libopenapi/datamodel/low/v3"
	"reflect"
)

type OperationChanges struct {
	PropertyChanges
	ExternalDocChanges         *ExternalDocChanges
	ParameterChanges           []*ParameterChanges
	ResponsesChanges           *ResponsesChanges
	SecurityRequirementChanges *SecurityRequirementChanges

	// v3
	RequestBodyChanges *RequestBodyChanges
	ServerChanges      []*ServerChanges
	ExtensionChanges   *ExtensionChanges
	// todo: callbacks need implementing.
	//CallbackChanges map[string]*CallbackChanges
}

func (o *OperationChanges) TotalChanges() int {
	c := o.PropertyChanges.TotalChanges()
	if o.ExternalDocChanges != nil {
		c += o.ExternalDocChanges.TotalChanges()
	}
	for k := range o.ParameterChanges {
		c += o.ParameterChanges[k].TotalChanges()
	}
	if o.ResponsesChanges != nil {
		c += o.RequestBodyChanges.TotalChanges()
	}
	if o.SecurityRequirementChanges != nil {
		c += o.SecurityRequirementChanges.TotalChanges()
	}
	if o.RequestBodyChanges != nil {
		c += o.RequestBodyChanges.TotalChanges()
	}
	for k := range o.ServerChanges {
		c += o.ServerChanges[k].TotalChanges()
	}
	// todo: add callbacks in here.
	if o.ExtensionChanges != nil {
		c += o.ExtensionChanges.TotalChanges()
	}
	return c
}

func (o *OperationChanges) TotalBreakingChanges() int {
	c := o.PropertyChanges.TotalBreakingChanges()
	if o.ExternalDocChanges != nil {
		c += o.ExternalDocChanges.TotalBreakingChanges()
	}
	for k := range o.ParameterChanges {
		c += o.ParameterChanges[k].TotalBreakingChanges()
	}
	if o.ResponsesChanges != nil {
		c += o.RequestBodyChanges.TotalBreakingChanges()
	}
	if o.SecurityRequirementChanges != nil {
		c += o.SecurityRequirementChanges.TotalBreakingChanges()
	}
	if o.RequestBodyChanges != nil {
		c += o.RequestBodyChanges.TotalBreakingChanges()
	}
	for k := range o.ServerChanges {
		c += o.ServerChanges[k].TotalBreakingChanges()
	}
	// todo: add callbacks in here.
	return c
}

func addSharedOperationProperties(left, right low.SharedOperations, changes *[]*Change) []*PropertyCheck {
	var props []*PropertyCheck

	// tags
	if len(left.GetTags().Value) > 0 || len(right.GetTags().Value) > 0 {
		ExtractStringValueSliceChanges(left.GetTags().Value, right.GetTags().Value,
			changes, v3.TagsLabel, false)
	}

	// summary
	addPropertyCheck(&props, left.GetSummary().ValueNode, right.GetSummary().ValueNode,
		left.GetSummary(), right.GetSummary(), changes, v3.SummaryLabel, false)

	// description
	addPropertyCheck(&props, left.GetDescription().ValueNode, right.GetDescription().ValueNode,
		left.GetDescription(), right.GetDescription(), changes, v3.DescriptionLabel, false)

	// deprecated
	addPropertyCheck(&props, left.GetDeprecated().ValueNode, right.GetDeprecated().ValueNode,
		left.GetDeprecated(), right.GetDeprecated(), changes, v3.DeprecatedLabel, false)

	return props
}

func compareSharedOperationObjects(l, r low.SharedOperations, changes *[]*Change, opChanges *OperationChanges) {

	// external docs
	if !l.GetExternalDocs().IsEmpty() && !r.GetExternalDocs().IsEmpty() {
		lExtDoc := l.GetExternalDocs().Value.(*base.ExternalDoc)
		rExtDoc := r.GetExternalDocs().Value.(*base.ExternalDoc)
		if !low.AreEqual(lExtDoc, rExtDoc) {
			opChanges.ExternalDocChanges = CompareExternalDocs(lExtDoc, rExtDoc)
		}
	}
	if l.GetExternalDocs().IsEmpty() && !r.GetExternalDocs().IsEmpty() {
		CreateChange(changes, ObjectAdded, v3.ExternalDocsLabel,
			nil, r.GetExternalDocs().ValueNode, false, nil,
			r.GetExternalDocs().Value)
	}
	if !l.GetExternalDocs().IsEmpty() && r.GetExternalDocs().IsEmpty() {
		CreateChange(changes, ObjectRemoved, v3.ExternalDocsLabel,
			l.GetExternalDocs().ValueNode, nil, false, l.GetExternalDocs().Value,
			nil)
	}

	// responses
	if !l.GetResponses().IsEmpty() && !r.GetResponses().IsEmpty() {
		opChanges.ResponsesChanges = CompareResponses(l, r)
	}
	if l.GetResponses().IsEmpty() && !r.GetResponses().IsEmpty() {
		CreateChange(changes, ObjectAdded, v3.ResponsesLabel,
			nil, r.GetResponses().ValueNode, false, nil,
			r.GetResponses().Value)
	}
	if !l.GetResponses().IsEmpty() && r.GetResponses().IsEmpty() {
		CreateChange(changes, ObjectRemoved, v3.ResponsesLabel,
			l.GetResponses().ValueNode, nil, true, l.GetResponses().Value,
			nil)
	}

}

func CompareOperations(l, r any) *OperationChanges {

	var changes []*Change
	var props []*PropertyCheck

	oc := new(OperationChanges)

	// Swagger
	if reflect.TypeOf(&v2.Operation{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v2.Operation{}) == reflect.TypeOf(r) {

		lOperation := l.(*v2.Operation)
		rOperation := r.(*v2.Operation)

		// perform hash check to avoid further processing
		if low.AreEqual(lOperation, rOperation) {
			return nil
		}

		props = append(props, addSharedOperationProperties(lOperation, rOperation, &changes)...)

		compareSharedOperationObjects(lOperation, rOperation, &changes, oc)

		// parameters
		lParamsUntyped := lOperation.GetParameters()
		rParamsUntyped := rOperation.GetParameters()
		if !lParamsUntyped.IsEmpty() && !rParamsUntyped.IsEmpty() {
			lParams := lParamsUntyped.Value.([]low.ValueReference[*v2.Parameter])
			rParams := rParamsUntyped.Value.([]low.ValueReference[*v2.Parameter])

			lv := make(map[string]*v2.Parameter, len(lParams))
			rv := make(map[string]*v2.Parameter, len(rParams))

			for i := range lParams {
				s := lParams[i].Value.Name.Value
				lv[s] = lParams[i].Value
			}
			for i := range rParams {
				s := rParams[i].Value.Name.Value
				rv[s] = rParams[i].Value
			}

			var paramChanges []*ParameterChanges
			for n := range lv {
				if _, ok := rv[n]; ok {
					if !low.AreEqual(lv[n], rv[n]) {
						ch := CompareParameters(lv[n], rv[n])
						if ch != nil {
							paramChanges = append(paramChanges, ch)
						}
					}
					continue
				}
				CreateChange(&changes, ObjectRemoved, v3.ParametersLabel,
					lv[n].Name.ValueNode, nil, true, lv[n].Name.Value,
					nil)

			}
			for n := range rv {
				if _, ok := lv[n]; !ok {
					CreateChange(&changes, ObjectAdded, v3.ParametersLabel,
						nil, rv[n].Name.ValueNode, true, nil,
						rv[n].Name.Value)
				}
			}
			oc.ParameterChanges = paramChanges
		}
	}

	// OpenAPI
	if reflect.TypeOf(&v3.Operation{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v3.Operation{}) == reflect.TypeOf(r) {

		lOperation := l.(*v3.Operation)
		rOperation := r.(*v3.Operation)

		// perform hash check to avoid further processing
		if low.AreEqual(lOperation, rOperation) {
			return nil
		}

		props = append(props, addSharedOperationProperties(lOperation, rOperation, &changes)...)

	}
	CheckProperties(props)
	oc.Changes = changes
	return oc
}
