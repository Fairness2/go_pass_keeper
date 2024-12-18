package payloads

type UpdateFile struct {
	ID      int64  `json:"id,omitempty"`
	Name    []byte `json:"name"`
	Comment string `json:"comment"`
}

type FileWithComment struct {
	ID      int64  `json:"id"`
	Name    []byte `json:"name"`
	Comment string `json:"comment"`
}
