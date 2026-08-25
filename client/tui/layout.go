package tui

type layout struct {
	sidebarWidth   int
	viewportWidth  int
	viewportHeight int
}

func computeLayout(width, height, inputRows int) layout {
	sidebar := width / 5

	switch {
	case sidebar < 20:
		sidebar = 20
	case sidebar > 32:
		sidebar = 32
	}

	const (
		viewportFrameX = 4
		chromeRows     = 8
		pillReserve    = 1
		sidebarBorder  = 2
	)

	return layout{
		sidebarWidth:   sidebar,
		viewportWidth:  width - sidebar - sidebarBorder - viewportFrameX,
		viewportHeight: height - chromeRows - pillReserve - inputRows,
	}
}

func (m *Model) resizePanes(width, height, inputRows int) {
	if width <= 0 || height <= 0 {
		return
	}

	m.ui.layout = computeLayout(width, height, inputRows)
	m.chat.viewport.Width = m.ui.layout.viewportWidth
	m.chat.viewport.Height = m.ui.layout.viewportHeight
	m.chat.input.SetWidth(m.ui.layout.viewportWidth)

	m.refreshViewport()
}
