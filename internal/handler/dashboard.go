package handler

import (
	"net/http"

	"github.com/templui/goilerplate/internal/ui"
	"github.com/templui/goilerplate/internal/ui/pages"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	ui.Render(w, r, pages.Dashboard())
}
