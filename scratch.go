package main
import (
	"fmt"
	"encoding/json"
)
type AuthorizeCommandPayload struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	Token   string `json:"card_token"`
}
type AuthorizeCommand struct {
	OrderID string
	Amount  int64
	Token   string
}
func main() {
	body := []byte(`{"order_id":"123","amount":100,"card_token":"tok_visa_success"}`)
	var payload AuthorizeCommandPayload
	json.Unmarshal(body, &payload)
	cmd := AuthorizeCommand(payload)
	fmt.Printf("Token: %q\n", cmd.Token)
}
