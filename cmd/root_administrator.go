package cmd

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func ensureRootAdministrator(c *core.CliContext, config *settings.Config) error {
	if config.RootAdminEmail == "" {
		return nil
	}

	rootAdministrator, err := services.Users.GetUserByEmail(c, config.RootAdminEmail)

	if err == errs.ErrUserNotFound {
		if config.RootAdminUsername == "" || config.RootAdminPassword == "" {
			log.BootWarnf(c, "[root_administrator.ensureRootAdministrator] root administrator is not created because username or password is not configured")
			return nil
		}

		rootAdministrator = &models.User{
			Username:              config.RootAdminUsername,
			Email:                 config.RootAdminEmail,
			Nickname:              "Root Administrator",
			Password:              config.RootAdminPassword,
			DefaultCurrency:       "VND",
			FirstDayOfWeek:        core.WEEKDAY_SUNDAY,
			TransactionEditScope:  models.TRANSACTION_EDIT_SCOPE_ALL,
			FeatureRestriction:    config.DefaultFeatureRestrictions,
			EmailVerified:         true,
			IsAdministrator:       true,
			IsRootAdministrator:   true,
		}

		if err = services.Users.CreateUser(c, rootAdministrator, false); err != nil {
			return err
		}

		log.BootInfof(c, "[root_administrator.ensureRootAdministrator] root administrator account was created")
		return nil
	}

	if err != nil {
		return err
	}

	if rootAdministrator.IsAdministrator && rootAdministrator.IsRootAdministrator {
		return nil
	}

	if err = services.Users.UpdateUserAdministratorStatus(c, rootAdministrator.Username, true, true); err != nil {
		return err
	}

	log.BootInfof(c, "[root_administrator.ensureRootAdministrator] existing root administrator account was protected")
	return nil
}
