package openapi

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"os"
	"testing"
)

func TestOpenAPIStandardValidation(t *testing.T) {
	data, e := os.ReadFile("../../../Docs/api.yaml")
	if e != nil {
		t.Fatal(e)
	}
	loader := openapi3.NewLoader()
	doc, e := loader.LoadFromData(data)
	if e != nil {
		t.Fatal(e)
	}
	if e = doc.Validate(context.Background()); e != nil {
		t.Fatal(e)
	}
}
