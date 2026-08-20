package tui

func (m Model) renderAuth() string {
	var errorLine string

	if m.auth.error != "" {
		errorLine = errorStyle.Render(
			"Error: " + m.auth.error + "\n\n",
		)
	}

	if m.auth.stage == stageServerPassword {
		m.auth.input.Prompt = "🔒 "

		content := authTitleStyle.Render("🔒 Server Password Required") +
			"\n\n" +
			"This server is locked. Enter the password to continue:" +
			"\n\n" +
			authInputStyle.Render(m.auth.input.View())

		return errorLine + "\n\n" + content
	}

	if m.auth.stage == stageMenu {
		return errorLine + "\n\n" + m.auth.menu.View()
	}

	if m.auth.stage == stageAuthenticating {
		return errorLine + "\n\n" + "Authenticating...\n\n"
	}

	label := authTitleStyle.Render(m.auth.choice + " - username:")

	if m.auth.stage == stagePassword {
		label = authTitleStyle.Render(m.auth.choice + " - password:")
	}

	return errorLine +
		"\n\n" +
		label +
		"\n\n" +
		authInputStyle.Render(m.auth.input.View())
}
