package payloads

// SaveCard представляет сохраненные данные карты вместе с комментариями пользователя.
type SaveCard struct {
	ID      int64  `json:"id,omitempty"`
	Number  []byte `json:"number"`
	Date    []byte `json:"date"`
	Owner   []byte `json:"owner"`
	CVV     []byte `json:"cvv"`
	Comment string `json:"comment"`
}

// Card представляет собой карту пользователя с минимальной конфиденциальной информацией, необходимой для безопасной обработки.
type Card struct {
	ID     int64  `json:"id"`
	Number []byte `json:"number"`
	Date   []byte `json:"date"`
	Owner  []byte `json:"owner"`
	CVV    []byte `json:"cvv"`
}

// CardWithComment представляет собой карточку пользователя, дополненную полем комментариев для дополнительного контекста или примечаний.
type CardWithComment struct {
	Card
	Comment string `json:"comment"`
}
