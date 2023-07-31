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

// BalanceChargedInput struct for data needed when charging balance
type BalanceChargedInput struct {
	Balance uint64 `json:"balance" binding:"required"`
}

// ChargeBalanceInput struct for data needed when charging balance
type ChargeBalanceInput struct {
	Balance int64 `json:"balance" binding:"required"`

	SuccessURL string `json:"success_url" binding:"required"`
	FailedURL  string `json:"failure_url" binding:"required"`
}

// BuyPackageInput for data needed when buying package
type BuyPackageInput struct {
	Vms           int           `json:"vms"  binding:"required"`
	PublicIPs     int           `json:"public_ips"  binding:"required"`
	VMType        models.VMType `json:"vm_type" binding:"required"`
	PeriodInMonth int           `json:"period"  binding:"required"`
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
		SuccessURL: stripe.String(input.SuccessURL),
		CancelURL:  stripe.String(input.FailedURL),
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

	balance, err := a.db.GetBalanceByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user balance is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	var balanceInUSD uint64
	if balance.Leftover > 0 {
		if balance.Leftover >= input.Balance {
			balance.Leftover -= input.Balance
		} else {
			balanceInUSD = input.Balance - balance.Leftover
			balance.Leftover = 0
		}
	}

	balance.BalanceInUSD += balanceInUSD
	err = a.db.UpdateBalance(balance)
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

	res := a.activatePackage(userID, input.VMType, input.Vms, input.PublicIPs, input.PeriodInMonth, false)
	if res != nil {
		return nil, res
	}

	return ResponseMsg{
		Message: "Package is bought successfully",
		Data:    nil,
	}, Ok()
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

	pkg, err := a.db.GetPackage(input.ID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("package is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	balance, err := a.db.GetBalanceByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	if balance.BalanceInUSD < pkg.Cost {
		return nil, BadRequest(errors.New("balance is not enough, please recharge your balance"))
	}

	pkg.PeriodInMonth *= 2
	err = a.db.UpdatePackage(pkg)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	switch pkg.VMType {
	case models.Small:
		balance.SmallVMsWithPublicIP += pkg.PublicIPs
		balance.SmallVMs += pkg.Vms - pkg.PublicIPs
	case models.Medium:
		balance.MediumVMsWithPublicIP += pkg.PublicIPs
		balance.MediumVMs += pkg.Vms - pkg.PublicIPs
	case models.Large:
		balance.LargeVMsWithPublicIP += pkg.PublicIPs
		balance.LargeVMs += pkg.Vms - pkg.PublicIPs
	}

	balance.BalanceInUSD -= pkg.Cost
	err = a.db.UpdateBalance(balance)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: "Package is renewed successfully",
		Data:    nil,
	}, Ok()
}

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

func (a *App) activatePackage(userID string, vmType models.VMType, vms, publicIPs, periodInMonth int, free bool) Response {
	if vms < publicIPs {
		return BadRequest(errors.New("virtual machines must be greater than or equal public ips"))
	}

	pkgRealCost, pkgCost, err := a.calculatePackageCost(vms, publicIPs, periodInMonth, vmType)
	if err != nil {
		log.Error().Err(err).Send()
		return InternalServerError(errors.New(internalServerErrorMsg))
	}

	pkg := models.Package{
		UserID:        userID,
		Vms:           vms,
		PublicIPs:     publicIPs,
		PeriodInMonth: periodInMonth,
		Cost:          pkgCost,
		RealCost:      pkgRealCost,
		CreatedAt:     time.Now(),
		VMType:        vmType,
	}

	balance, err := a.db.GetBalanceByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		return NotFound(errors.New("user is not found"))
	}
	if err != nil {
		log.Error().Err(err).Send()
		return InternalServerError(errors.New(internalServerErrorMsg))
	}

	if balance.BalanceInUSD < pkgCost && !free {
		return BadRequest(errors.New("balance is not enough, please recharge your balance"))
	}

	err = a.db.CreatePackage(&pkg)
	if err != nil {
		log.Error().Err(err).Send()
		return InternalServerError(errors.New(internalServerErrorMsg))
	}

	switch vmType {
	case models.Small:
		balance.SmallVMsWithPublicIP += publicIPs
		balance.SmallVMs += vms - publicIPs
	case models.Medium:
		balance.MediumVMsWithPublicIP += publicIPs
		balance.MediumVMs += vms - publicIPs
	case models.Large:
		balance.LargeVMsWithPublicIP += publicIPs
		balance.LargeVMs += vms - publicIPs
	}

	if !free {
		balance.BalanceInUSD -= pkgCost
	}

	err = a.db.UpdateBalance(balance)
	if err != nil {
		log.Error().Err(err).Send()
		return InternalServerError(errors.New(internalServerErrorMsg))
	}

	return nil
}

func (a *App) calculatePackageCost(vms, publicIPs, periodInMonth int, vmType models.VMType) (uint64, uint64, error) {
	var vmCPU, vmMemory, vmDisk, vmCost, vmCostWithPublicIP uint64
	switch vmType {
	case models.Small:
		vmCPU = SmallCPU
		vmMemory = SmallMemory
		vmDisk = SmallDisk
		vmCost = a.config.Prices.SmallVM
		vmCostWithPublicIP = a.config.Prices.SmallVMWithPublicIP
	case models.Medium:
		vmCPU = MediumCPU
		vmMemory = MediumMemory
		vmDisk = MediumDisk
		vmCost = a.config.Prices.MediumVM
		vmCostWithPublicIP = a.config.Prices.MediumVMWithPublicIP
	case models.Large:
		vmCPU = LargeCPU
		vmMemory = LargeMemory
		vmDisk = LargeDisk
		vmCost = a.config.Prices.LargeVM
		vmCostWithPublicIP = a.config.Prices.LargeVMWithPublicIP
	}

	var pkgRealCost, pkgCost uint64
	for i := 1; i <= vms; i++ {
		publicIP := publicIPs > 0
		cost, err := a.calculator.CalculateCost(int64(vmCPU)*int64(vms), int64(vmMemory)*int64(vms), 0, int64(vmDisk)*int64(vms), publicIP, false)
		if err != nil {
			return 0, 0, err
		}

		if publicIPs > 0 {
			pkgCost += vmCostWithPublicIP
		} else {
			pkgCost += vmCost
		}

		pkgRealCost += uint64(cost * 1e7)
		publicIPs--
	}

	pkgRealCost = pkgRealCost * uint64(periodInMonth)
	pkgCost = pkgCost * uint64(periodInMonth)
	return pkgRealCost, pkgCost, nil
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

			balance, err := a.db.GetBalanceByUserID(pkg.UserID)
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

				// add a daily leftover
				balance.Leftover += pkg.Cost / uint64(30*pkg.PeriodInMonth)
				err = a.db.UpdateBalance(balance)
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
