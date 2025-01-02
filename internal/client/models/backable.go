package models

import tea "github.com/charmbracelet/bubbletea"

// Backable — это структура, инкапсулирующая предыдущую tea.Model, позволяющая вернуться к предыдущей модели.
type Backable struct {
	lastModel tea.Model
}

// Back возвращает последнюю сохраненную tea.Model и вызывает ее метод Init для повторной инициализации.
func (b Backable) Back() (tea.Model, tea.Cmd) {
	l := b.lastModel
	return l, l.Init()
}

// NewBackable создает и возвращает экземпляр Backable, инкапсулирующий предыдущую модель для целей навигации.
func NewBackable(lastModel tea.Model) Backable {
	return Backable{lastModel: lastModel}
}
