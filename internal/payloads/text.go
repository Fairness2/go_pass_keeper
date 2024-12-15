package payloads

type SaveText struct {
	ID       int64  `json:"id,omitempty"`
	TextData []byte `json:"text_data"`
	Comment  string `json:"comment"`
}

type Text struct {
	ID       int64  `json:"id"`
	TextData []byte `json:"text_data"`
}

type TextWithComment struct {
	Text
	Comment string `json:"comment"`
}
