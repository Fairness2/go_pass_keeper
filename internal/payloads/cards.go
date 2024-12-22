package payloads

import "passkeeper/internal/encrypt/cipher"

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

func (item *CardWithComment) Encrypt(ch *cipher.Cipher) error {
	eNumber, err := ch.Encrypt(item.Number)
	if err != nil {
		return err
	}
	item.Number = eNumber
	if len(item.Date) > 0 {
		eDate, err := ch.Encrypt(item.Date)
		if err != nil {
			return err
		}
		item.Date = eDate
	}
	if len(item.Owner) > 0 {
		eOwner, err := ch.Encrypt(item.Owner)
		if err != nil {
			return err
		}
		item.Owner = eOwner
	}
	if len(item.CVV) > 0 {
		eCVV, err := ch.Encrypt(item.CVV)
		if err != nil {
			return err
		}
		item.CVV = eCVV
	}
	return nil
}

func (item *CardWithComment) Decrypt(ch *cipher.Cipher) error {
	dNumber, err := ch.Decrypt(item.Number)
	if err != nil {
		return err
	}
	item.Number = dNumber
	if len(item.Date) > 0 {
		eDate, err := ch.Decrypt(item.Date)
		if err != nil {
			return err
		}
		item.Date = eDate
	}
	if len(item.Owner) > 0 {
		eOwner, err := ch.Decrypt(item.Owner)
		if err != nil {
			return err
		}
		item.Owner = eOwner
	}
	if len(item.CVV) > 0 {
		eCVV, err := ch.Decrypt(item.CVV)
		if err != nil {
			return err
		}
		item.CVV = eCVV
	}
	return nil
}
