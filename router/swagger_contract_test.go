package router

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	apidocs "github.com/zHElEARN/go-csust-planet/docs"
)

const errorResponseRef = "#/definitions/response.ErrorResponse"

type swaggerDocument struct {
	Paths       map[string]map[string]swaggerOperation `json:"paths"`
	Definitions map[string]swaggerDefinition           `json:"definitions"`
}

type swaggerOperation struct {
	Description string                     `json:"description"`
	Responses   map[string]swaggerResponse `json:"responses"`
}

type swaggerResponse struct {
	Schema struct {
		Ref string `json:"$ref"`
	} `json:"schema"`
}

type swaggerDefinition struct {
	Properties map[string]json.RawMessage `json:"properties"`
}

func TestSwaggerPreservesExistingHTTPContract(t *testing.T) {
	expected := map[string][]string{
		"GET /admin/announcements":                         {"200", "401", "500"},
		"POST /admin/announcements":                        {"201", "400", "401", "500"},
		"GET /admin/announcements/{id}":                    {"200", "400", "401", "404", "500"},
		"PUT /admin/announcements/{id}":                    {"200", "400", "401", "404", "500"},
		"DELETE /admin/announcements/{id}":                 {"204", "400", "401", "404", "500"},
		"GET /admin/app-versions":                          {"200", "401", "500"},
		"POST /admin/app-versions":                         {"201", "400", "401", "409", "500"},
		"GET /admin/app-versions/{id}":                     {"200", "400", "401", "404", "500"},
		"PUT /admin/app-versions/{id}":                     {"200", "400", "401", "404", "409", "500"},
		"DELETE /admin/app-versions/{id}":                  {"204", "400", "401", "404", "500"},
		"GET /admin/semester-calendars":                    {"200", "401", "500"},
		"POST /admin/semester-calendars":                   {"201", "400", "401", "409", "500"},
		"GET /admin/semester-calendars/{semester_code}":    {"200", "401", "404", "500"},
		"PUT /admin/semester-calendars/{semester_code}":    {"200", "400", "401", "404", "409", "500"},
		"DELETE /admin/semester-calendars/{semester_code}": {"204", "401", "404", "500"},
		"GET /config/announcements":                        {"200", "500"},
		"GET /config/app-versions":                         {"200", "400", "500"},
		"GET /config/app-versions/check":                   {"200", "400", "500"},
		"GET /config/campus-map":                           {"200", "500"},
		"GET /config/semester-calendars":                   {"200", "500"},
		"GET /config/semester-calendars/{semester_code}":   {"200", "400", "404"},
	}

	var document swaggerDocument
	if err := json.Unmarshal([]byte(apidocs.SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatalf("failed to parse generated Swagger document: %v", err)
	}

	errorDefinition, ok := document.Definitions["response.ErrorResponse"]
	if !ok {
		t.Fatal("expected response.ErrorResponse definition")
	}
	if _, ok := errorDefinition.Properties["error"]; !ok {
		t.Fatal("expected response.ErrorResponse to contain the error property")
	}

	for operationKey, expectedStatuses := range expected {
		method, path, ok := strings.Cut(operationKey, " ")
		if !ok {
			t.Fatalf("invalid expected operation key %q", operationKey)
		}
		operation, ok := document.Paths[path][strings.ToLower(method)]
		if !ok {
			t.Errorf("expected Swagger operation %s", operationKey)
			continue
		}
		if operation.Description == "" {
			t.Errorf("expected Swagger operation %s to have a description", operationKey)
		}

		actualStatuses := make([]string, 0, len(operation.Responses))
		for status, response := range operation.Responses {
			actualStatuses = append(actualStatuses, status)
			if !strings.HasPrefix(status, "2") && response.Schema.Ref != errorResponseRef {
				t.Errorf("expected %s response %s to reference %s, got %q", operationKey, status, errorResponseRef, response.Schema.Ref)
			}
		}
		slices.Sort(actualStatuses)
		slices.Sort(expectedStatuses)
		if !slices.Equal(actualStatuses, expectedStatuses) {
			t.Errorf("unexpected responses for %s: got %v, want %v", operationKey, actualStatuses, expectedStatuses)
		}
	}
}
