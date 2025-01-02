package models

import tea "github.com/charmbracelet/bubbletea"

type Backable struct {
	lastModel tea.Model
}

func (b Backable) Back() (tea.Model, tea.Cmd) {
	l := b.lastModel
	return l, l.Init()
}

func NewBackable(lastModel tea.Model) Backable {
	return Backable{lastModel: lastModel}
}
