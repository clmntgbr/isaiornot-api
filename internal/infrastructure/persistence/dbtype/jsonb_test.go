package dbtype_test

import (
	"testing"

	"go-api/internal/infrastructure/persistence/dbtype"
)

func TestJSONBValueReturnsString(t *testing.T) {
	payload := dbtype.JSONB(`{"firstName":"Clément"}`)
	value, err := payload.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	got, ok := value.(string)
	if !ok {
		t.Fatalf("Value() type = %T, want string", value)
	}
	if got != `{"firstName":"Clément"}` {
		t.Fatalf("Value() = %q, want JSON string", got)
	}
}
