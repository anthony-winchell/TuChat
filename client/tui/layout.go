package tui

type layout struct {
	sidebarWidth   int
	viewportWidth  int
	viewportHeight int
}

func computeLayout(width, height int) layout {
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
		viewportHeight: height - chromeRows - pillReserve,
	}
}
