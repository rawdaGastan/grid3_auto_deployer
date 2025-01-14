package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type AddCardInput struct {
	TokenID   string `json:"token_id" binding:"required" validate:"nonzero"`
	TokenType string `json:"token_type" binding:"required" validate:"nonzero"`
}

type SetDefaultCardInput struct {
	PaymentMethodID string `json:"payment_method_id" binding:"required" validate:"nonzero"`
}

type ChargeBalance struct {
	PaymentMethodID string  `json:"payment_method_id" binding:"required" validate:"nonzero"`
	Amount          float64 `json:"amount" binding:"required" validate:"nonzero"`
}

// Example endpoint: Add a new card
// @Summary Add a new card
// @Description Add a new card
// @Tags Card
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param card body AddCardInput true "Card input"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/card [post]
func (a *App) AddCardHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input AddCardInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read input data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid input data"))
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// if user has no stipe customer ID then we create it
	if len(strings.TrimSpace(user.StripeCustomerID)) == 0 {
		customer, err := createCustomer(user.Name(), user.Email)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		user.StripeCustomerID = customer.ID
		err = a.db.UpdateUserByID(user)
		if err == gorm.ErrRecordNotFound {
			return nil, NotFound(errors.New("user is not found"))
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	paymentMethod, err := createPaymentMethod(input.TokenType, input.TokenID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	unique, err := a.db.IsCardUnique(paymentMethod.Card.Fingerprint)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if !unique {
		return nil, BadRequest(errors.New("card is added before"))
	}

	err = attachPaymentMethod(user.StripeCustomerID, paymentMethod.ID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// Add payment method in DB
	if err := a.db.AddCard(
		&models.Card{
			UserID:          userID,
			PaymentMethodID: paymentMethod.ID,
			CustomerID:      user.StripeCustomerID,
			CardType:        input.TokenType,
			ExpMonth:        paymentMethod.Card.ExpMonth,
			ExpYear:         paymentMethod.Card.ExpYear,
			Last4:           paymentMethod.Card.Last4,
			Brand:           string(paymentMethod.Card.Brand),
			Fingerprint:     paymentMethod.Card.Fingerprint,
		},
	); err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// if no payment is added before then we update the user payment ID with it as a default
	if len(strings.TrimSpace(user.StripeDefaultPaymentID)) == 0 {
		// Update the default payment method for future payments
		err = updateDefaultPaymentMethod(user.StripeCustomerID, paymentMethod.ID)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		user.StripeDefaultPaymentID = paymentMethod.ID
		err = a.db.UpdateUserByID(user)
		if err == gorm.ErrRecordNotFound {
			return nil, NotFound(errors.New("user is not found"))
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	// try to settle old invoices using the card
	invoices, err := a.db.ListUnpaidInvoices(user.ID.String())
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	for _, invoice := range invoices {
		response := a.payInvoice(&user, paymentMethod.ID, voucherAndBalanceAndCard, invoice.Total, invoice.ID)
		if response.Err() != nil {
			log.Error().Err(response.Err()).Send()
		}
	}

	return ResponseMsg{
		Message: "Card is added successfully",
		Data:    nil,
	}, Created()
}

// Example endpoint: Set card as default
// @Summary Set card as default
// @Description Set card as default
// @Tags Card
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param card body SetDefaultCardInput true "Card input"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/card/default [put]
func (a *App) SetDefaultCardHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input SetDefaultCardInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read input data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid input data"))
	}

	card, err := a.db.GetCardByPaymentMethod(input.PaymentMethodID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("card is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = attachPaymentMethod(card.CustomerID, card.PaymentMethodID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// Update the default payment method for future payments
	err = updateDefaultPaymentMethod(card.CustomerID, input.PaymentMethodID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	err = a.db.UpdateUserByID(models.User{ID: uuid.MustParse(userID), StripeDefaultPaymentID: card.PaymentMethodID})
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Card is set as default successfully",
		Data:    nil,
	}, Created()
}

// Example endpoint: List user's cards
// @Summary List user's cards
// @Description List user's cards
// @Tags Card
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.Card
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/card [get]
func (a *App) ListCardHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	cards, err := a.db.GetUserCards(userID)
	if err == gorm.ErrRecordNotFound || len(cards) == 0 {
		return ResponseMsg{
			Message: "no cards found",
			Data:    cards,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Cards are found",
		Data:    cards,
	}, Ok()
}

// Example endpoint: Delete user card
// @Summary Delete user card
// @Description Delete user card
// @Tags Card
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /user/card/{id} [delete]
func (a *App) DeleteCardHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read card id"))
	}

	card, err := a.db.GetCard(id)
	if err == gorm.ErrRecordNotFound {
		return nil, BadRequest(errors.New("card is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if userID != card.UserID {
		return nil, NotFound(errors.New("card is not found"))
	}

	cards, err := a.db.GetUserCards(userID)
	if err == gorm.ErrRecordNotFound || len(cards) == 0 {
		return ResponseMsg{
			Message: "No cards found",
			Data:    nil,
		}, Ok()
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	// check active deployments
	var vms []models.VM
	var k8s []models.K8sCluster
	if len(cards) == 1 {
		vms, err = a.db.GetAllSuccessfulVms(userID)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		k8s, err = a.db.GetAllSuccessfulK8s(userID)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	if len(vms) > 0 && len(k8s) > 0 {
		return nil, BadRequest(errors.New("you have active deployment and cannot delete the card"))
	}

	// Update the default payment method for future payments (if deleted card is the default)
	if card.PaymentMethodID == user.StripeDefaultPaymentID {
		var newPaymentMethod string
		// no more cards
		if len(cards) == 1 {
			newPaymentMethod = ""
		}

		for _, c := range cards {
			if c.PaymentMethodID != user.StripeDefaultPaymentID {
				newPaymentMethod = c.PaymentMethodID
				if err = updateDefaultPaymentMethod(card.CustomerID, c.PaymentMethodID); err != nil {
					log.Error().Err(err).Send()
					return nil, InternalServerError(errors.New(internalServerErrorMsg))
				}
				break
			}
		}

		err = a.db.UpdateUserPaymentMethod(userID, newPaymentMethod)
		if err == gorm.ErrRecordNotFound {
			return nil, NotFound(errors.New("user is not found"))
		}
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	// If user has another cards or no active deployments, so can delete
	err = detachPaymentMethod(card.PaymentMethodID)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if err = a.db.DeleteCard(id); err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Card is deleted successfully",
		Data:    nil,
	}, Ok()
}
