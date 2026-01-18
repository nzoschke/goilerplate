package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/templui/goilerplate/internal/service"
	"github.com/templui/goilerplate/internal/ui"
	"github.com/templui/goilerplate/internal/ui/blocks"
	"github.com/templui/goilerplate/internal/ui/pages"
	"github.com/templui/goilerplate/internal/validation"
)

type newsletterHandler struct {
	emailService *service.EmailService
}

func NewNewsletterHandler(emailService *service.EmailService) *newsletterHandler {
	return &newsletterHandler{
		emailService: emailService,
	}
}

func (h *newsletterHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	isHomeForm := r.Header.Get("HX-Target") == "home-newsletter"

	err := validation.ValidateEmail(email)
	if err != nil {
		if isHomeForm {
			ui.Render(w, r, pages.HomeNewsletterSection("Please provide a valid email address"))
		} else {
			ui.Render(w, r, blocks.FooterNewsletterForm("Please provide a valid email address"))
		}
		return
	}

	err = h.emailService.SubscribeNewsletter(email)
	if err != nil {
		// Service layer already logs errors - just handle error case
		// Return success to prevent email enumeration
		slog.Warn("newsletter subscription error", "error", err, "email", email)
	}

	// Always show success (prevents email enumeration)
	if isHomeForm {
		ui.Render(w, r, pages.HomeNewsletterSuccess())
	} else {
		ui.Render(w, r, blocks.FooterNewsletterSuccess())
	}
}