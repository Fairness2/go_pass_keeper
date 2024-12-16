package payloads

type SaveCard struct {
	ID      int64  `json:"id,omitempty"`
	Number  []byte `json:"number"`
	Date    []byte `json:"date"`
	Owner   []byte `json:"owner"`
	CVV     []byte `json:"cvv"`
	Comment string `json:"comment"`
}

type Card struct {
	ID     int64  `json:"id"`
	Number []byte `json:"number"`
	Date   []byte `json:"date"`
	Owner  []byte `json:"owner"`
	CVV    []byte `json:"cvv"`
}

type CardWithComment struct {
	Card
	Comment string `json:"comment"`
}
