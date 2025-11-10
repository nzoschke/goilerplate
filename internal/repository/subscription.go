package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/templui/goilerplate/internal/model"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type SubscriptionRepository interface {
	Create(sub *model.Subscription) error
	ByUserID(userID string) (*model.Subscription, error)
	ByProviderSubscriptionID(providerSubID string) (*model.Subscription, error)
	Update(sub *model.Subscription) error
}

type subscriptionRepository struct {
	db *sqlx.DB
}

func NewSubscriptionRepository(db *sqlx.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, plan_id, status,
			provider_customer_id, provider_subscription_id,
			current_period_end, amount, currency, interval,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(
		query,
		sub.ID,
		sub.UserID,
		sub.PlanID,
		sub.Status,
		sub.ProviderCustomerID,
		sub.ProviderSubscriptionID,
		sub.CurrentPeriodEnd,
		sub.Amount,
		sub.Currency,
		sub.Interval,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	return err
}

func (r *subscriptionRepository) ByUserID(userID string) (*model.Subscription, error) {
	sub := &model.Subscription{}
	query := `SELECT * FROM subscriptions WHERE user_id = $1`

	err := r.db.Get(sub, query, userID)
	if err == sql.ErrNoRows {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *subscriptionRepository) ByProviderSubscriptionID(providerSubID string) (*model.Subscription, error) {
	sub := &model.Subscription{}
	query := `SELECT * FROM subscriptions WHERE provider_subscription_id = $1`

	err := r.db.Get(sub, query, providerSubID)
	if err == sql.ErrNoRows {
		return nil, ErrSubscriptionNotFound
	}
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *subscriptionRepository) Update(sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET plan_id = $1,
		    status = $2,
		    provider_customer_id = $3,
		    provider_subscription_id = $4,
		    current_period_end = $5,
		    amount = $6,
		    currency = $7,
		    interval = $8,
		    updated_at = $9
		WHERE id = $10
	`

	result, err := r.db.Exec(
		query,
		sub.PlanID,
		sub.Status,
		sub.ProviderCustomerID,
		sub.ProviderSubscriptionID,
		sub.CurrentPeriodEnd,
		sub.Amount,
		sub.Currency,
		sub.Interval,
		sub.UpdatedAt,
		sub.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}
