package tui

func (m Model) renderAuth() string {
	var errorLine string

	if m.auth.error != "" {
		errorLine = errorStyle.Render(
			"Error: " + m.auth.error + "\n\n",
		)
	}

	if m.auth.stage == stageServerPassword {
		return errorLine + "This server requires a password\nEnter it below:\n\n" +
			m.auth.input.View()
	}

	if m.auth.stage == stageMenu {
		return errorLine + m.auth.menu.View()
	}

	if m.auth.stage == stageAuthenticating {
		return errorLine + "Authenticating...\n\n"
	}

	label := m.auth.choice + " - username:"

	if m.auth.stage == stagePassword {
		label = m.auth.choice + " - password:"
	}

	return errorLine +
		label +
		"\n\n" +
		m.auth.input.View()
}
