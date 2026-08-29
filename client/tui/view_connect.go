package tui

func (m Model) renderConnect() string {
	var errorLine string
	if m.connect.error != "" {
		errorLine = errorStyle.Render("Error: " + m.connect.error + "\n\n")
	}

	content := authTitleStyle.Render("🔗 Connect to Server") +
		"\n\n" +
		"Enter the server's host:port, or skip this screen" +
		"\n" +
		"with -addr or $TUCHAT_ADDR." +
		"\n\n" +
		authInputStyle.Render(m.connect.input.View())

	return errorLine + "\n\n" + content
}
