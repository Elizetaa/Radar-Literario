package openapi

import (
	"encoding/json"
	"testing"
)

func TestValidateRejectsDanglingReferenceAndUndeclaredPathParameter(t *testing.T) {
	base := object{"openapi": "3.0.3", "paths": object{"/works/{id}": object{"get": object{"operationId": "get_work", "responses": object{"200": object{"description": "ok"}}}}}, "components": object{"schemas": object{}}}
	data, _ := json.Marshal(base)
	if Validate(data) == nil {
		t.Fatal("undeclared path parameter accepted")
	}
	op := base["paths"].(object)["/works/{id}"].(object)["get"].(object)
	op["parameters"] = []any{object{"name": "id", "in": "path", "required": true, "schema": idSchema()}}
	op["bad"] = ref("Missing")
	data, _ = json.Marshal(base)
	if Validate(data) == nil {
		t.Fatal("dangling reference accepted")
	}
	delete(op, "bad")
	data, _ = json.Marshal(base)
	if err := Validate(data); err != nil {
		t.Fatal(err)
	}
}
