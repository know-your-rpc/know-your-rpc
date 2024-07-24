package queries

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestHelloName(t *testing.T) {
	address, err := ExtractSigner("0x9403bbd93cd0be364e1e20ee568d5ea5c03e78034fad7728bd6aba2efc5ce49436d2f7faf752ae6a9aa6cf0f085427325c65d19bfc2bde1767be231d29557c401b", "action=authorize_all version=0 domain=localhost valid_until=1722285352")

	assert.Equal(t, address, "0x3d62C20583AefDAe7959bad67D457e6D24d7A656")
	assert.Equal(t, err, nil)
}
