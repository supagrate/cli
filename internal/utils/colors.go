package utils

import "github.com/charmbracelet/lipgloss"

func Red(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(str)
}

func Green(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(str)
}

func Yellow(str string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(str)
}
