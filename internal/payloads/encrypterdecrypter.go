package payloads

// Encrypter — это интерфейс, определяющий метод шифрования фрагментов байтов и возврата зашифрованного результата или ошибки.
type Encrypter interface {
	Encrypt(body []byte) ([]byte, error)
}

// Decrypter определяет поведение расшифровки данного фрагмента байта, возвращая расшифрованные данные или ошибку.
type Decrypter interface {
	Decrypt(body []byte) ([]byte, error)
}
