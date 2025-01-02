package validators

import "github.com/asaskevich/govalidator"

// init регистрирует пользовательские валидаторы в govalidator.
func init() {
	govalidator.CustomTypeTagMap.Set("requireByteArray", arrayNotEmpty)
}

// arrayNotEmpty проверяет, является ли ввод непустым байтовым срезом. Возвращает true, если не пусто, в противном случае — false.
func arrayNotEmpty(i any, _ any) bool {
	switch v := i.(type) {
	case []byte:
		return len(v) > 0
	}
	return false
}
