package models

// IncrementCircleIndex циклически корректирует значение индекса на основе заданного ключа и общего количества полей.
// Если клавиша «вверх» или «shift+tab», индекс уменьшается, в противном случае — увеличивается.
// Результат циклически изменяется в диапазоне от 0 до fieldLen включительно.
func IncrementCircleIndex(index, fieldsLen int, key string) int {
	// Cycle indexes
	if key == "up" || key == "shift+tab" {
		index--
	} else {
		index++
	}
	if index > fieldsLen {
		index = 0
	} else if index < 0 {
		index = fieldsLen
	}
	return index
}
