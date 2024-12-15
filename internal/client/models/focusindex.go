package models

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
