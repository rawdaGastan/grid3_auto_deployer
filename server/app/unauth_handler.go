package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetMaintenanceHandler gets maintenance flag
// Example endpoint: Gets maintenance flag
// @Summary Gets maintenance flag
// @Description Gets maintenance flag
// @Tags Unauthorized/Authorized
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Maintenance
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /maintenance [get]
func (a *App) GetMaintenanceHandler(req *http.Request) (interface{}, Response) {
	maintenance, err := a.db.GetMaintenance()
	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("maintenance is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: fmt.Sprintf("Maintenance is set with %v", maintenance.Active),
		Data:    maintenance,
	}, Ok()
}

// GetNextLaunchHandler returns next launch state
// Example endpoint: Gets next launch state
// @Summary Gets next launch state
// @Description Gets next launch state
// @Tags Unauthorized/Authorized
// @Accept  json
// @Produce  json
// @Success 200 {object} models.NextLaunch
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /nextlaunch [get]
func (a *App) GetNextLaunchHandler(req *http.Request) (interface{}, Response) {
	nextLaunch, err := a.db.GetNextLaunch()

	if err == gorm.ErrRecordNotFound {
		return nil, NotFound(errors.New("next launch is not found"))
	}

	if err != nil {
		log.Error().Err(err).Send()
		return nil, InternalServerError(errors.New(internalServerErrorMsg))
	}

	return ResponseMsg{
		Message: fmt.Sprintf("Next Launch is Launched with state: %v", nextLaunch.Launched),
		Data:    nextLaunch,
	}, Ok()
}
