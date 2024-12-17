package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codescalers/cloud4students/deployer"
	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type method string

const (
	card                     method = "card"
	balance                  method = "balance"
	voucher                  method = "voucher"
	voucherAndBalance        method = "voucher+balance"
	voucherAndCard           method = "voucher+card"
	balanceAndCard           method = "balance+card"
	voucherAndBalanceAndCard method = "voucher+balance+card"
)

var methods = []method{
	card, balance, voucher,
	voucherAndBalance, voucherAndCard, balanceAndCard,
	voucherAndBalanceAndCard,
}

type PayInvoiceInput struct {
	Method        method `json:"method" binding:"required"`
	CardPaymentID string `json:"card_payment_id"`
}

// ListInvoicesHandler lists user's invoices
// Example endpoint: Lists user's invoices
// @Summary Lists user's invoices
// @Description Lists user's invoices
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.Invoice
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /invoice [get]
func (a *App) ListInvoicesHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	invoices, err := a.db.ListUserInvoices(userID)
	if err == gorm.ErrRecordNotFound || len(invoices) == 0 {
		return ResponseMsg{
			Message: "no invoices found",
			Data:    invoices,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Invoices are found",
		Data:    invoices,
	}, Ok()
}

// GetInvoiceHandler gets user's invoice by ID
// Example endpoint: Gets user's invoice by ID
// @Summary Gets user's invoice by ID
// @Description Gets user's invoice by ID
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} models.Invoice
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /invoice/{id} [get]
func (a *App) GetInvoiceHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read invoice id"))
	}

	invoice, err := a.db.GetInvoice(id)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("invoice is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if userID != invoice.UserID {
		return nil, NotFound(errors.New("invoice is not found"))
	}

	return ResponseMsg{
		Message: "Invoice exists",
		Data:    invoice,
	}, Ok()
}

// PayInvoiceHandler pay user's invoice
// Example endpoint: Pay user's invoice
// @Summary Pay user's invoice
// @Description Pay user's invoice
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Param payment body PayInvoiceInput true "Payment method and ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /invoice/pay/{id} [put]
func (a *App) PayInvoiceHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	id, err := strconv.Atoi(mux.Vars(req)["id"])
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to invoice card id"))
	}

	var input PayInvoiceInput
	err = json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read input data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid input data"))
	}

	invoice, err := a.db.GetInvoice(id)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("invoice is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if userID != invoice.UserID {
		return nil, NotFound(errors.New("invoice is not found"))
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var paymentDetails models.PaymentDetails

	switch input.Method {
	case card:
		_, err := createPaymentIntent(user.StripeCustomerID, input.CardPaymentID, a.config.Currency, invoice.Total)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, BadRequest(errors.New("payment failed, please try again later or report the problem"))
		}

		paymentDetails = models.PaymentDetails{Card: invoice.Total}

	case balance:
		if user.Balance < invoice.Total {
			return nil, BadRequest(errors.New("balance is not enough to pay the invoice"))
		}

		paymentDetails = models.PaymentDetails{Balance: invoice.Total}

		user.Balance -= invoice.Total
		if err = a.db.UpdateUserByID(user); err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

	case voucher:
		if user.VoucherBalance < invoice.Total {
			return nil, BadRequest(errors.New("voucher balance is not enough to pay the invoice"))
		}

		paymentDetails = models.PaymentDetails{VoucherBalance: invoice.Total}

		user.VoucherBalance -= invoice.Total
		if err = a.db.UpdateUserByID(user); err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

	case voucherAndBalance:
		if user.VoucherBalance+user.Balance < invoice.Total {
			return nil, BadRequest(errors.New("voucher balance and balance are not enough to pay the invoice"))
		}

		if user.VoucherBalance > invoice.Total {
			paymentDetails = models.PaymentDetails{VoucherBalance: invoice.Total}
			user.VoucherBalance -= invoice.Total
		} else {
			paymentDetails = models.PaymentDetails{VoucherBalance: user.VoucherBalance, Balance: (invoice.Total - user.VoucherBalance)}
			user.Balance = (invoice.Total - user.VoucherBalance)
			user.VoucherBalance = 0
		}

		if err = a.db.UpdateUserByID(user); err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

	case voucherAndCard:
		if user.VoucherBalance > invoice.Total {
			paymentDetails = models.PaymentDetails{VoucherBalance: invoice.Total}
			user.VoucherBalance -= invoice.Total
		} else {
			paymentDetails = models.PaymentDetails{VoucherBalance: user.VoucherBalance, Card: (invoice.Total - user.VoucherBalance)}
			_, err := createPaymentIntent(user.StripeCustomerID, input.CardPaymentID, a.config.Currency, invoice.Total-user.VoucherBalance)
			if err != nil {
				log.Error().Err(err).Send()
				return nil, BadRequest(errors.New("payment failed, please try again later or report the problem"))
			}
			user.VoucherBalance = 0
		}

		if err = a.db.UpdateUserByID(user); err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

	case balanceAndCard:
		if user.Balance > invoice.Total {
			paymentDetails = models.PaymentDetails{Balance: invoice.Total}
			user.Balance -= invoice.Total
		} else {
			_, err := createPaymentIntent(user.StripeCustomerID, input.CardPaymentID, a.config.Currency, invoice.Total-user.Balance)
			if err != nil {
				log.Error().Err(err).Send()
				return nil, BadRequest(errors.New("payment failed, please try again later or report the problem"))
			}
			paymentDetails = models.PaymentDetails{Balance: user.Balance, Card: (invoice.Total - user.Balance)}
			user.Balance = 0
		}

		if err = a.db.UpdateUserByID(user); err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

	case voucherAndBalanceAndCard:
		if user.VoucherBalance > invoice.Total {
			paymentDetails = models.PaymentDetails{Balance: invoice.Total}
			user.VoucherBalance -= invoice.Total
		} else if user.Balance+user.VoucherBalance > invoice.Total {
			paymentDetails = models.PaymentDetails{VoucherBalance: user.VoucherBalance, Balance: (invoice.Total - user.VoucherBalance)}
			user.Balance = (invoice.Total - user.VoucherBalance)
			user.VoucherBalance = 0
		} else {
			_, err := createPaymentIntent(user.StripeCustomerID, input.CardPaymentID, a.config.Currency, invoice.Total-user.VoucherBalance-user.Balance)
			if err != nil {
				log.Error().Err(err).Send()
				return nil, BadRequest(errors.New("payment failed, please try again later or report the problem"))
			}
			paymentDetails = models.PaymentDetails{
				Balance: user.Balance, VoucherBalance: user.VoucherBalance,
				Card: (invoice.Total - user.Balance - user.VoucherBalance),
			}
			user.VoucherBalance = 0
			user.Balance = 0
		}

		if err = a.db.UpdateUserByID(user); err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

	default:
		return nil, BadRequest(fmt.Errorf("invalid payment method, only methods allowed %v", methods))
	}

	err = a.db.PayInvoice(id, paymentDetails)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("invoice is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Invoice is paid successfully",
		Data:    nil,
	}, Ok()
}

func (a *App) monthlyInvoices() {
	for {
		now := time.Now()
		monthLastDay := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
		timeTilLast := monthLastDay.Sub(now)

		if now.Day() != monthLastDay.Day() {
			// Wait until the last day of the month
			time.Sleep(timeTilLast)
		}

		users, err := a.db.ListAllUsers()
		if err == gorm.ErrRecordNotFound || len(users) == 0 {
			log.Error().Err(err).Msg("Users are not found")
		}

		if err != nil {
			log.Error().Err(err).Send()
		}

		// TODO: what if routine is killed
		// Create invoices for all system users
		for _, user := range users {
			// 1. Create new monthly invoice
			if err = a.createInvoice(user.ID.String(), now); err != nil {
				log.Error().Err(err).Send()
			}

			// 2. Use balance/voucher balance to pay invoices
			user.Balance, user.VoucherBalance, err = a.db.PayUserInvoices(user.ID.String(), user.Balance, user.VoucherBalance)
			if err != nil {
				log.Error().Err(err).Send()
			} else {
				if err = a.db.UpdateUserByID(user); err != nil {
					log.Error().Err(err).Send()
				}
			}

			// 3. Use cards to pay invoices
			if err = a.payUserInvoicesUsingCards(user.ID.String(), user.StripeCustomerID, user.StripeDefaultPaymentID, true); err != nil {
				log.Error().Err(err).Send()
			}

			// 4. Delete expired deployments with invoices not paid for more than 3 months
			if err = a.deleteInvoiceDeploymentsNotPaidSince3Months(user.ID.String(), now); err != nil {
				log.Error().Err(err).Send()
			}
		}

		// Calculate the next last day of the month
		nextMonthLastDay := time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
		timeTilNextLast := nextMonthLastDay.Sub(now)

		// Wait until the last day of the next month
		time.Sleep(timeTilNextLast)
	}
}

func (a *App) createInvoice(userID string, now time.Time) error {
	usagePercentageInMonth := deployer.UsagePercentageInMonth(now)
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.Local)

	vms, err := a.db.GetAllVms(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	k8s, err := a.db.GetAllK8s(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	var items []models.DeploymentItem
	var total float64

	for _, vm := range vms {
		cost := float64(vm.PricePerMonth) * usagePercentageInMonth

		items = append(items, models.DeploymentItem{
			DeploymentResources: vm.Resources,
			DeploymentType:      "vm",
			DeploymentID:        vm.ID,
			HasPublicIP:         vm.Public,
			PeriodInHours:       time.Since(firstDayOfMonth).Hours(),
			Cost:                cost,
		})

		total += cost
	}

	for _, cluster := range k8s {
		cost := float64(cluster.PricePerMonth) * usagePercentageInMonth

		items = append(items, models.DeploymentItem{
			DeploymentResources: cluster.Master.Resources,
			DeploymentType:      "k8s",
			DeploymentID:        cluster.ID,
			HasPublicIP:         cluster.Master.Public,
			PeriodInHours:       time.Since(firstDayOfMonth).Hours(),
			Cost:                cost,
		})

		total += cost
	}

	if err = a.db.CreateInvoice(&models.Invoice{
		UserID:      userID,
		Total:       total,
		Deployments: items,
	}); err != nil {
		return err
	}

	return nil
}

// payUserInvoicesUsingCards tries to pay invoices with user cards
func (a *App) payUserInvoicesUsingCards(userID, customerID, defaultPaymentMethod string, useOtherCards bool) error {
	// get unpaid invoices
	invoices, err := a.db.ListUnpaidInvoices(userID)
	if err != nil {
		return err
	}

	cards, err := a.db.GetUserCards(userID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		// 1. use default payment method
		if len(defaultPaymentMethod) != 0 {
			_, err := createPaymentIntent(customerID, defaultPaymentMethod, a.config.Currency, invoice.Total)
			if err != nil {
				log.Error().Err(err).Send()
			} else {
				if err := a.db.PayInvoice(invoice.ID, models.PaymentDetails{Card: invoice.Total}); err != nil {
					log.Error().Err(err).Send()
				}
				continue
			}
		}

		if !useOtherCards {
			continue
		}

		// 2. check other user cards
		for _, card := range cards {
			if defaultPaymentMethod != card.PaymentMethodID {
				_, err := createPaymentIntent(customerID, card.PaymentMethodID, a.config.Currency, invoice.Total)
				if err != nil {
					log.Error().Err(err).Send()
				} else {
					if err := a.db.PayInvoice(invoice.ID, models.PaymentDetails{Card: invoice.Total}); err != nil {
						log.Error().Err(err).Send()
					}
					break
				}
			}
		}
	}

	return nil
}

func (a *App) deleteInvoiceDeploymentsNotPaidSince3Months(userID string, now time.Time) error {
	invoices, err := a.db.ListUnpaidInvoices(userID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		threeMonthsAgo := now.AddDate(0, -3, 0)

		// check if the invoice created 3 months ago (not after it) and not paid
		if !invoice.CreatedAt.After(threeMonthsAgo) && !invoice.Paid {
			for _, dl := range invoice.Deployments {
				if dl.DeploymentType == "vm" {
					if err = a.db.DeleteVMByID(dl.DeploymentID); err != nil {
						log.Error().Err(err).Send()
					}
				}

				if dl.DeploymentType == "k8s" {
					if err = a.db.DeleteK8s(dl.DeploymentID); err != nil {
						log.Error().Err(err).Send()
					}
				}
			}
		}
	}

	return nil
}

func (a *App) sendRemindersToPayInvoices() {
	ticker := time.NewTicker(time.Hour * 24)

	for range ticker.C {
		now := time.Now()

		users, err := a.db.ListAllUsers()
		if err != nil {
			log.Error().Err(err).Send()
		}

		for _, u := range users {
			if err = a.sendInvoiceReminderToUser(u.ID.String(), u.Email, u.Name(), now); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}
}

func (a *App) sendInvoiceReminderToUser(userID, userEmail, userName string, now time.Time) error {
	invoices, err := a.db.ListUnpaidInvoices(userID)
	if err != nil {
		return err
	}

	currencyName, err := getCurrencyName(a.config.Currency)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		oneMonthsAgo := now.AddDate(0, -1, 0)
		oneWeekAgo := now.AddDate(0, 0, -7)

		// check if the invoice created 1 months ago (not after it) and
		// last remainder sent for this invoice was 7 days ago and
		// invoice is not paid
		if invoice.CreatedAt.Before(oneMonthsAgo) &&
			invoice.LastReminderAt.Before(oneWeekAgo) &&
			!invoice.Paid {
			// overdue date starts after one month since invoice creation
			overDueStart := invoice.CreatedAt.AddDate(0, 1, 0)
			overDueDays := int(now.Sub(overDueStart).Hours() / 24)

			// 3 months as a grace period
			deadline := invoice.CreatedAt.AddDate(0, 3, 0)
			gracePeriod := int(deadline.Sub(now).Hours() / 24)

			mailBody := "We hope this message finds you well.\n"
			mailBody += fmt.Sprintf("Our records show that there is an outstanding invoice for %v %s associated with your account (%d). ", invoice.Total, currencyName, invoice.ID)
			mailBody += fmt.Sprintf("As of today, the payment for this invoice is %d days overdue.", overDueDays)
			mailBody += "To avoid any interruptions to your services and the potential deletion of your deployments, "
			mailBody += fmt.Sprintf("we kindly ask that you make the payment within the next %d days. If the invoice remains unpaid after this period, ", gracePeriod)
			mailBody += "please be advised that the associated deployments will be deleted from our system.\n\n"

			mailBody += "You can easily pay your invoice by charging balance, activating voucher or using cards.\n\n"
			mailBody += "If you have already made the payment or need any assistance, "
			mailBody += "please don't hesitate to reach out to us.\n\n"
			mailBody += "We appreciate your prompt attention to this matter and thank you fosr being a valued customer."

			subject := "Unpaid Invoice Notification – Action Required"
			subject, body := internal.AdminMailContent(subject, mailBody, a.config.Server.Host, userName)

			if err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, userEmail, subject, body); err != nil {
				log.Error().Err(err).Send()
			}

			notification := models.Notification{UserID: userID, Msg: fmt.Sprintf("Reminder: %s", mailBody)}
			err = a.db.CreateNotification(&notification)
			if err != nil {
				log.Error().Err(err).Send()
			}

			if err = a.db.UpdateInvoiceLastRemainderDate(invoice.ID); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}

	return nil
}

// getCurrencyName returns the full name of the currency based on the currency code.
func getCurrencyName(currencyCode string) (string, error) {
	currencyMap := map[string]string{
		"USD": "US Dollar",
		"EUR": "Euro",
		"GBP": "British Pound",
		"AUD": "Australian Dollar",
		"CAD": "Canadian Dollar",
		"JPY": "Japanese Yen",
		"CNY": "Chinese Yuan",
		"INR": "Indian Rupee",
		"MXN": "Mexican Peso",
		"BRL": "Brazilian Real",
		"RUB": "Russian Ruble",
		"KRW": "South Korean Won",
		"CHF": "Swiss Franc",
		"SEK": "Swedish Krona",
		"NZD": "New Zealand Dollar",
	}

	currencyCode = strings.ToUpper(currencyCode)

	if currencyName, exists := currencyMap[currencyCode]; exists {
		return currencyName, nil
	}

	return "", errors.New("unknown currency")
}
