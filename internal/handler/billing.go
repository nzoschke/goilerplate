package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/templui/goilerplate/internal/ctxkeys"
	"github.com/templui/goilerplate/internal/service"
	"github.com/templui/goilerplate/internal/ui"
	"github.com/templui/goilerplate/internal/ui/pages"
)

type BillingHandler struct {
	subscriptionService *service.SubscriptionService
	polarService        *service.PolarService
}

func NewBillingHandler(subscriptionService *service.SubscriptionService, polarService *service.PolarService) *BillingHandler {
	return &BillingHandler{
		subscriptionService: subscriptionService,
		polarService:        polarService,
	}
}

func (h *BillingHandler) BillingPage(w http.ResponseWriter, r *http.Request) {
	ui.Render(w, r, pages.Billing())
}

func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	user := ctxkeys.User(r.Context())

	profile := ctxkeys.Profile(r.Context())
	if profile == nil {
		http.Error(w, "Profile not found", http.StatusInternalServerError)
		return
	}

	planID := r.FormValue("plan_id")
	if planID == "" {
		http.Error(w, "Invalid plan selected", http.StatusBadRequest)
		return
	}

	interval := r.FormValue("interval")
	if interval == "" {
		interval = "monthly"
	}

	checkoutURL, err := h.polarService.CreateCheckoutURL(user.ID, planID, interval, user.Email, profile.Name)
	if err != nil {
		slog.Error("failed to create checkout", "error", err, "user_id", user.ID, "plan_id", planID)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	slog.Info("redirecting to polar checkout", "user_id", user.ID, "checkout_url", checkoutURL)
	http.Redirect(w, r, checkoutURL, http.StatusSeeOther)
}

func (h *BillingHandler) PolarWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read webhook payload", "error", err)
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}
	defer func() {
		closeErr := r.Body.Close()
		if closeErr != nil {
			slog.Error("failed to close request body", "error", closeErr)
		}
	}()

	webhookID := r.Header.Get("webhook-id")
	timestamp := r.Header.Get("webhook-timestamp")
	signature := r.Header.Get("webhook-signature")

	err = h.polarService.HandleWebhook(payload, webhookID, timestamp, signature)
	if err != nil {
		slog.Error("failed to handle webhook", "error", err)
		http.Error(w, "Failed to process webhook", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received": true}`))
}

func (h *BillingHandler) CustomerPortal(w http.ResponseWriter, r *http.Request) {
	user := ctxkeys.User(r.Context())

	portalURL, err := h.polarService.CustomerPortalURL(user.ID)
	if err != nil {
		slog.Error("failed to get customer portal", "error", err, "user_id", user.ID)
		http.Error(w, "Failed to access customer portal", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, portalURL, http.StatusSeeOther)
}
