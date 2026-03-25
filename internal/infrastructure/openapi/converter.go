package openapi

import (
	"encoding/json"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/nuvivo314/mcp-executor/internal/domain/model"
)

// convertEndpoints converts an openapi3.T document into a flat list of domain Endpoints.
func convertEndpoints(doc *openapi3.T) []model.Endpoint {
	var endpoints []model.Endpoint

	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			if op == nil {
				continue
			}
			ep := model.Endpoint{
				Path:        path,
				Method:      strings.ToUpper(method),
				OperationID: op.OperationID,
				Summary:     op.Summary,
				Description: op.Description,
			}

			for _, pRef := range op.Parameters {
				if pRef == nil || pRef.Value == nil {
					continue
				}
				p := pRef.Value
				param := model.Parameter{
					Name:     p.Name,
					In:       p.In,
					Required: p.Required,
				}
				if p.Schema != nil && p.Schema.Value != nil {
					if raw, err := json.Marshal(p.Schema.Value); err == nil {
						param.Schema = string(raw)
					}
				}
				ep.Parameters = append(ep.Parameters, param)
			}

			if op.RequestBody != nil && op.RequestBody.Value != nil {
				rb := op.RequestBody.Value
				ep.RequestBody = &model.RequestBody{
					Required: rb.Required,
				}
				for contentType, mt := range rb.Content {
					ep.RequestBody.ContentType = contentType
					if mt.Schema != nil && mt.Schema.Value != nil {
						if raw, err := json.Marshal(mt.Schema.Value); err == nil {
							ep.RequestBody.Schema = string(raw)
						}
					}
					break // take first content type
				}
			}

			endpoints = append(endpoints, ep)
		}
	}

	return endpoints
}
