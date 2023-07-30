// Package app for c4s backend app
package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/codescalers/cloud4students/internal"
	"github.com/codescalers/cloud4students/middlewares"
	"github.com/codescalers/cloud4students/models"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

var (
	vmCPU    = uint64(1)
	vmMemory = uint64(2)
	vmDisk   = uint64(25)
)

//TODO: vouchers and quota

// BalanceChargedInput struct for data needed when charging balance
type BalanceChargedInput struct {
	Balance float64 `json:"balance" binding:"required"`
}

// ChargeBalanceInput struct for data needed when charging balance
type ChargeBalanceInput struct {
	Balance int64 `json:"balance" binding:"required"`

	SuccessUrl string `json:"success_url" binding:"required"`
	FailedUrl  string `json:"failure_url" binding:"required"`
}

// BuyPackageInput for data needed when buying package
type BuyPackageInput struct {
	Vms           int `json:"vms"  binding:"required"`
	PublicIPs     int `json:"public_ips"  binding:"required"`
	PeriodInMonth int `json:"period"  binding:"required"`
}

// RenewPackageInput for data needed when renewing package
type RenewPackageInput struct {
	ID int `json:"id"  binding:"required"`
}

func (a *App) chargeBalanceHandler(req *http.Request) (interface{}, Response) {
	var input ChargeBalanceInput
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

	priceID, err := createBalanceProductInStripe(input.Balance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	paramsCheckout := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(input.SuccessUrl),
		CancelURL:  stripe.String(input.FailedUrl),
	}

	s, err := session.New(paramsCheckout)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Redirect",
		Data:    s.URL,
	}, Ok()
}

func (a *App) balanceChargedHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input BalanceChargedInput
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("failed to read data"))
	}

	err = validator.Validate(input)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, BadRequest(errors.New("invalid data"))
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var balance float64
	if user.LeftoverBalance > 0 {
		if user.LeftoverBalance >= input.Balance {
			user.LeftoverBalance -= input.Balance
		} else {
			balance = input.Balance - user.LeftoverBalance
			user.LeftoverBalance = 0
		}
	}

	user.Balance += balance
	err = a.db.UpdateUserByID(user)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Balance is updated successfully",
		Data:    nil,
	}, Ok()
}

func (a *App) buyPackageHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input BuyPackageInput
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

	if input.Vms < input.PublicIPs {
		return nil, BadRequest(errors.New("virtual machines must be greater than public ips"))
	}

	var pkgCost float64
	for i := 1; i <= input.Vms; i++ {
		publicIP := input.PublicIPs > 0
		cost, err := a.calculator.CalculateCost(int64(vmCPU)*int64(input.Vms), int64(vmMemory)*int64(input.Vms), 0, int64(vmDisk)*int64(input.Vms), publicIP, false)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, InternalServerError(errors.New(internalServerErrorMsg))
		}

		pkgCost += cost
		input.PublicIPs--
	}

	pkgCost = pkgCost * float64(input.PeriodInMonth)

	pkg := models.Package{
		UserID:        userID,
		Vms:           input.Vms,
		PublicIPs:     input.PublicIPs,
		PeriodInMonth: input.PeriodInMonth,
		Cost:          pkgCost,
		CreatedAt:     time.Now(),
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if user.Balance < pkgCost {
		return nil, BadRequest(errors.New("balance is not enough, please recharge your balance"))
	}

	err = a.db.CreatePackage(&pkg)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	user.Balance -= pkgCost
	err = a.db.UpdateUserByID(user)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Package is bought successfully",
		Data:    nil,
	}, Ok()
}

// ListPackagesHandler returns all packages of user
func (a *App) listPackagesHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	packages, err := a.db.ListPackages(userID)
	if err == gorm.ErrRecordNotFound || len(packages) == 0 {
		return ResponseMsg{
			Message: "no packages found",
			Data:    packages,
		}, Ok()
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Packages are found",
		Data:    packages,
	}, Ok()
}

func (a *App) notifyUsersExpiredPackages() {
	ticker := time.NewTicker(24 * time.Hour * time.Duration(a.config.NotifyUsersExpirationInDays))

	for range ticker.C {
		packages, err := a.db.GetExpiredPackages(a.config.ExpirationToleranceInDays)
		if err != nil {
			log.Error().Err(err).Send()
		}

		for _, pkg := range packages {
			user, err := a.db.GetUserByID(pkg.UserID)
			if err != nil {
				log.Error().Err(err).Send()
			}

			expiredAt := time.Since(pkg.CreatedAt.AddDate(0, pkg.PeriodInMonth, 0))
			daysLeft := a.config.ExpirationToleranceInDays - int(expiredAt)

			if daysLeft > 0 {
				subject, body := internal.NotifyExpiredPackages(daysLeft, a.config.Server.Host)

				err = internal.SendMail(a.config.MailSender.Email, a.config.MailSender.SendGridKey, user.Email, subject, body)
				if err != nil {
					log.Error().Err(err).Send()
				}
				continue
			}

			// delete expired vms
			vms, err := a.db.GetExpiredVms(user.ID.String())
			if err != nil {
				log.Error().Err(err).Send()
			}

			for _, vm := range vms {
				err = a.db.DeleteVMByID(vm.ID)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}

			// delete expired clusters
			clusters, err := a.db.GetExpiredK8s(user.ID.String())
			if err != nil {
				log.Error().Err(err).Send()
			}

			for _, k8s := range clusters {
				err = a.db.DeleteK8s(k8s.ID)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	}
}

func (a *App) renewPackageHandler(req *http.Request) (interface{}, Response) {
	userID := req.Context().Value(middlewares.UserIDKey("UserID")).(string)

	var input RenewPackageInput
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

	pkg, err := a.db.GetPkgByID(input.ID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("package is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	user, err := a.db.GetUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if user.Balance < pkg.Cost {
		return nil, BadRequest(errors.New("balance is not enough, please recharge your balance"))
	}

	err = a.db.UpdatePackage(pkg.ID, pkg.PeriodInMonth*2)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	user.Balance -= pkg.Cost
	err = a.db.UpdateUserByID(user)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Package is renewed successfully",
		Data:    nil,
	}, Ok()
}
