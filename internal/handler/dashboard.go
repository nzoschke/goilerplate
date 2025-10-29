package handler

import (
	"log/slog"
	"net/http"

	"github.com/templui/goilerplate/internal/ctxkeys"
	"github.com/templui/goilerplate/internal/service"
	"github.com/templui/goilerplate/internal/ui"
	"github.com/templui/goilerplate/internal/ui/pages"
)

type DashboardHandler struct {
	goalService *service.GoalService
}

func NewDashboardHandler(goalService *service.GoalService) *DashboardHandler {
	return &DashboardHandler{
		goalService: goalService,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	user := ctxkeys.User(r.Context())

	goals, err := h.goalService.Goals(user.ID, "recent")
	if err != nil {
		slog.Error("failed to get goals", "error", err, "user_id", user.ID)
		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
		return
	}

	ui.Render(w, r, pages.Dashboard(goals))
}
