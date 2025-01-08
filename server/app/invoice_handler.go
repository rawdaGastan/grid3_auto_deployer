package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	Method        method `json:"method" validate:"nonzero" binding:"required"`
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

// DownloadInvoiceHandler downloads user's invoice by ID
// Example endpoint: Downloads user's invoice by ID
// @Summary Downloads user's invoice by ID
// @Description Downloads user's invoice by ID
// @Tags Invoice
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /invoice/download/{id} [get]
func (a *App) DownloadInvoiceHandler(req *http.Request) (interface{}, Response) {
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

	// Get downloads dir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	downloadsDir := filepath.Join(homeDir, "Downloads")
	pdfPath := filepath.Join(downloadsDir, fmt.Sprintf("invoice-%s-%d.pdf", invoice.UserID, invoice.ID))

	err = os.WriteFile(pdfPath, invoice.FileData, 0644)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: fmt.Sprintf("Invoice is downloaded successfully at %s", pdfPath),
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

	response := a.payInvoice(&user, input.CardPaymentID, input.Method, invoice.Total, id)
	if response.Err() != nil {
		return nil, response
	}

	return ResponseMsg{
		Message: "Invoice is paid successfully",
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
			if err = a.createInvoice(user, now); err != nil {
				log.Error().Err(err).Send()
			}

			// 2. Pay invoices
			invoices, err := a.db.ListUnpaidInvoices(user.ID.String())
			if err != nil {
				log.Error().Err(err).Send()
			}

			for _, invoice := range invoices {
				cards, err := a.db.GetUserCards(user.ID.String())
				if err != nil {
					log.Error().Err(err).Send()
				}

				// No cards option
				if len(cards) == 0 {
					response := a.payInvoice(&user, "", voucherAndBalance, invoice.Total, invoice.ID)
					if response.Err() != nil {
						log.Error().Err(response.Err()).Send()
					}
					continue
				}

				// Use default card
				response := a.payInvoice(&user, user.StripeDefaultPaymentID, voucherAndBalanceAndCard, invoice.Total, invoice.ID)
				if response.Err() != nil {
					log.Error().Err(response.Err()).Send()
				} else {
					continue
				}

				for _, card := range cards {
					if card.PaymentMethodID == user.StripeDefaultPaymentID {
						continue
					}

					response := a.payInvoice(&user, card.PaymentMethodID, voucherAndBalanceAndCard, invoice.Total, invoice.ID)
					if response.Err() != nil {
						log.Error().Err(response.Err()).Send()
						continue
					}
					break
				}
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

func (a *App) createInvoice(user models.User, now time.Time) error {
	monthStart := time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.Local)

	vms, err := a.db.GetAllSuccessfulVms(user.ID.String())
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	k8s, err := a.db.GetAllSuccessfulK8s(user.ID.String())
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	var items []models.DeploymentItem
	var total float64

	for _, vm := range vms {
		usageStart := monthStart
		if vm.CreatedAt.After(monthStart) {
			usageStart = vm.CreatedAt
		}

		usagePercentageInMonth, err := deployer.UsagePercentageInMonth(usageStart, now)
		if err != nil {
			return err
		}

		cost := float64(vm.PricePerMonth) * usagePercentageInMonth

		items = append(items, models.DeploymentItem{
			DeploymentResources: vm.Resources,
			DeploymentType:      "vm",
			DeploymentID:        vm.ID,
			DeploymentName:      vm.Name,
			DeploymentCreatedAt: vm.CreatedAt,
			HasPublicIP:         vm.Public,
			PeriodInHours:       time.Since(usageStart).Hours(),
			Cost:                cost,
		})

		total += cost
	}

	for _, cluster := range k8s {
		usageStart := monthStart
		if cluster.CreatedAt.After(monthStart) {
			usageStart = cluster.CreatedAt
		}

		usagePercentageInMonth, err := deployer.UsagePercentageInMonth(usageStart, now)
		if err != nil {
			return err
		}

		cost := float64(cluster.PricePerMonth) * usagePercentageInMonth

		items = append(items, models.DeploymentItem{
			DeploymentResources: cluster.Master.Resources,
			DeploymentType:      "k8s",
			DeploymentID:        cluster.ID,
			DeploymentName:      cluster.Master.Name,
			DeploymentCreatedAt: cluster.CreatedAt,
			HasPublicIP:         cluster.Master.Public,
			PeriodInHours:       time.Since(usageStart).Hours(),
			Cost:                cost,
		})

		total += cost
	}

	if len(items) > 0 {
		in := models.Invoice{
			UserID:      user.ID.String(),
			Total:       total,
			Deployments: items,
		}

		// Creating pdf for invoice
		pdfContent, err := internal.CreateInvoicePDF(in, user)
		if err != nil {
			return err
		}

		in.FileData = pdfContent

		// Creating invoice in db
		if err = a.db.CreateInvoice(&in); err != nil {
			return err
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
		// oneMonthsAgo := now.AddDate(0, -1, 0)
		oneWeekAgo := now.AddDate(0, 0, -7)

		// check if the invoice created 1 months ago (not after it) and
		// last remainder sent for this invoice was before 7 days ago and
		// invoice is not paid
		// invoice.CreatedAt.Before(oneMonthsAgo) &&
		if invoice.LastReminderAt.Before(oneWeekAgo) &&
			!invoice.Paid {
			// overdue date starts after one month since invoice creation
			overDueStart := invoice.CreatedAt.AddDate(0, 1, 0)
			overDueDays := int(now.Sub(overDueStart).Hours() / 24)

			// 3 months as a grace period
			deadline := invoice.CreatedAt.AddDate(0, 3, 0)
			gracePeriod := int(deadline.Sub(now).Hours() / 24)

			mailBody := "We hope this message finds you well.\n"
			mailBody += fmt.Sprintf("Our records show that there is an outstanding invoice for %v %s associated with your account (%d). ", invoice.Total, currencyName, invoice.ID)
			if overDueDays > 0 {
				mailBody += fmt.Sprintf("As of today, the payment for this invoice is %d days overdue.", overDueDays)
			}
			mailBody += "To avoid any interruptions to your services and the potential deletion of your deployments, "
			mailBody += fmt.Sprintf("we kindly ask that you make the payment within the next %d days. If the invoice remains unpaid after this period, ", gracePeriod)
			mailBody += "please be advised that the associated deployments will be deleted from our system.\n\n"

			mailBody += "You can easily pay your invoice by charging balance, activating voucher or using cards.\n\n"
			mailBody += "If you have already made the payment or need any assistance, "
			mailBody += "please don't hesitate to reach out to us.\n\n"
			mailBody += "We appreciate your prompt attention to this matter and thank you for being a valued customer."

			subject := "Unpaid Invoice Notification – Action Required"
			subject, body := internal.AdminMailContent(subject, mailBody, a.config.Server.Host, userName)

			if err = internal.SendMail(
				a.config.MailSender.Email, a.config.MailSender.SendGridKey, userEmail, subject, body,
				fmt.Sprintf("invoice-%s-%d.pdf", invoice.UserID, invoice.ID), invoice.FileData,
			); err != nil {
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

func (a *App) pay(user *models.User, cardPaymentID string, method method, invoiceTotal float64) (models.PaymentDetails, error) {
	var paymentDetails models.PaymentDetails

	switch method {
	case card:
		_, err := createPaymentIntent(user.StripeCustomerID, cardPaymentID, a.config.Currency, invoiceTotal)
		if err != nil {
			log.Error().Err(err).Send()
			return paymentDetails, errors.New("payment failed, please try again later or report the problem")
		}

		paymentDetails = models.PaymentDetails{Card: invoiceTotal}

	case balance:
		if user.Balance < invoiceTotal {
			return paymentDetails, errors.New("balance is not enough to pay the invoice")
		}

		paymentDetails = models.PaymentDetails{Balance: invoiceTotal}
		user.Balance -= invoiceTotal

	case voucher:
		if user.VoucherBalance < invoiceTotal {
			return paymentDetails, errors.New("voucher balance is not enough to pay the invoice")
		}

		paymentDetails = models.PaymentDetails{VoucherBalance: invoiceTotal}
		user.VoucherBalance -= invoiceTotal

	case voucherAndBalance:
		if user.VoucherBalance+user.Balance < invoiceTotal {
			return paymentDetails, errors.New("voucher balance and balance are not enough to pay the invoice")
		}

		if user.VoucherBalance >= invoiceTotal {
			paymentDetails = models.PaymentDetails{VoucherBalance: invoiceTotal}
			user.VoucherBalance -= invoiceTotal
		} else {
			paymentDetails = models.PaymentDetails{VoucherBalance: user.VoucherBalance, Balance: (invoiceTotal - user.VoucherBalance)}
			user.Balance = (invoiceTotal - user.VoucherBalance)
			user.VoucherBalance = 0
		}

	case voucherAndCard:
		if user.VoucherBalance >= invoiceTotal {
			paymentDetails = models.PaymentDetails{VoucherBalance: invoiceTotal}
			user.VoucherBalance -= invoiceTotal
		} else {
			paymentDetails = models.PaymentDetails{VoucherBalance: user.VoucherBalance, Card: (invoiceTotal - user.VoucherBalance)}
			_, err := createPaymentIntent(user.StripeCustomerID, cardPaymentID, a.config.Currency, invoiceTotal-user.VoucherBalance)
			if err != nil {
				log.Error().Err(err).Send()
				return paymentDetails, errors.New("payment failed, please try again later or report the problem")
			}
			user.VoucherBalance = 0
		}

	case balanceAndCard:
		if user.Balance >= invoiceTotal {
			paymentDetails = models.PaymentDetails{Balance: invoiceTotal}
			user.Balance -= invoiceTotal
		} else {
			_, err := createPaymentIntent(user.StripeCustomerID, cardPaymentID, a.config.Currency, invoiceTotal-user.Balance)
			if err != nil {
				log.Error().Err(err).Send()
				return paymentDetails, errors.New("payment failed, please try again later or report the problem")
			}
			paymentDetails = models.PaymentDetails{Balance: user.Balance, Card: (invoiceTotal - user.Balance)}
			user.Balance = 0
		}

	case voucherAndBalanceAndCard:
		if user.VoucherBalance >= invoiceTotal {
			paymentDetails = models.PaymentDetails{Balance: invoiceTotal}
			user.VoucherBalance -= invoiceTotal

		} else if user.Balance+user.VoucherBalance >= invoiceTotal {
			paymentDetails = models.PaymentDetails{VoucherBalance: user.VoucherBalance, Balance: (invoiceTotal - user.VoucherBalance)}
			user.Balance = (invoiceTotal - user.VoucherBalance)
			user.VoucherBalance = 0

		} else {
			_, err := createPaymentIntent(user.StripeCustomerID, cardPaymentID, a.config.Currency, invoiceTotal-user.VoucherBalance-user.Balance)
			if err != nil {
				log.Error().Err(err).Send()
				return paymentDetails, errors.New("payment failed, please try again later or report the problem")
			}

			paymentDetails = models.PaymentDetails{
				Balance: user.Balance, VoucherBalance: user.VoucherBalance,
				Card: (invoiceTotal - user.Balance - user.VoucherBalance),
			}
			user.VoucherBalance = 0
			user.Balance = 0
		}

	default:
		return paymentDetails, fmt.Errorf("invalid payment method, only methods allowed %v", methods)
	}

	return paymentDetails, nil
}

func (a *App) payInvoice(user *models.User, cardPaymentID string, method method, invoiceTotal float64, invoiceID int) Response {
	paymentDetails, err := a.pay(user, cardPaymentID, method, invoiceTotal)
	if err != nil {
		return BadRequest(errors.New(internalServerErrorMsg))
	}

	// invoice used voucher balance
	if paymentDetails.VoucherBalance != 0 {
		if err = a.db.UpdateUserVoucherBalance(user.ID.String(), user.VoucherBalance); err != nil {
			log.Error().Err(err).Send()
			return InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	// invoice used balance
	if paymentDetails.Balance != 0 {
		if err = a.db.UpdateUserBalance(user.ID.String(), user.Balance); err != nil {
			log.Error().Err(err).Send()
			return InternalServerError(errors.New(internalServerErrorMsg))
		}
	}

	paymentDetails.InvoiceID = invoiceID
	err = a.db.PayInvoice(invoiceID, paymentDetails)
	if err == gorm.ErrRecordNotFound {
		return NotFound(errors.New("invoice is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return InternalServerError(errors.New(internalServerErrorMsg))
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
