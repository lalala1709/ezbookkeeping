package api

import (
	"crypto/subtle"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

const adminPasswordHeaderName = "X-EZB-Admin-Password"

// AdministrationApi provides limited, password-protected member administration.
type AdministrationApi struct {
	ApiUsingConfig
	users                   *services.UserService
	tokens                  *services.TokenService
	transactions            *services.TransactionService
	categories              *services.TransactionCategoryService
	tags                    *services.TransactionTagService
	tagGroups               *services.TransactionTagGroupService
	templates               *services.TransactionTemplateService
	userCustomIcons         *services.UserCustomIconService
	userCustomExchangeRates *services.UserCustomExchangeRatesService
	insightsExplorers       *services.InsightsExplorerService
}

// Initialize an administration api singleton instance.
var (
	Administration = &AdministrationApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		users:                   services.Users,
		tokens:                  services.Tokens,
		transactions:            services.Transactions,
		categories:              services.TransactionCategories,
		tags:                    services.TransactionTags,
		tagGroups:               services.TransactionTagGroups,
		templates:               services.TransactionTemplates,
		userCustomIcons:         services.UserCustomIcons,
		userCustomExchangeRates: services.UserCustomExchangeRates,
		insightsExplorers:       services.InsightsExplorers,
	}
)

// ListUsersHandler returns active member accounts.
func (a *AdministrationApi) ListUsersHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.authorize(c); err != nil {
		return nil, err
	}

	users, err := a.users.GetAllUsers(c)

	if err != nil {
		log.Errorf(c, "[administration.ListUsersHandler] failed to get all users, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	userInfos := make([]*models.AdminUserInfo, len(users))

	for i := 0; i < len(users); i++ {
		user := users[i]
		userInfos[i] = &models.AdminUserInfo{
			Username:        user.Username,
			Email:           user.Email,
			Nickname:        user.Nickname,
			Disabled:        user.Disabled,
			EmailVerified:   user.EmailVerified,
			CreatedUnixTime: user.CreatedUnixTime,
			LastLoginAt:     user.LastLoginUnixTime,
		}
	}

	return &models.AdminUserListResponse{
		TotalUserCount: int64(len(userInfos)),
		Users:          userInfos,
	}, nil
}

// UpdateUserPasswordHandler resets a member password and revokes their active sessions.
func (a *AdministrationApi) UpdateUserPasswordHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.authorize(c); err != nil {
		return nil, err
	}

	var request models.AdminUserPasswordUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Warnf(c, "[administration.UpdateUserPasswordHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	request.Username = strings.TrimSpace(request.Username)
	user, err := a.users.GetUserByUsername(c, request.Username)

	if err != nil {
		log.Warnf(c, "[administration.UpdateUserPasswordHandler] failed to get user \"%s\", because %s", request.Username, err.Error())
		return nil, errs.ErrUserNotFound
	}

	user.Password = request.Password
	err = a.users.UpdateUserPassword(c, user)

	if err != nil {
		log.Errorf(c, "[administration.UpdateUserPasswordHandler] failed to update password for user \"%s\", because %s", request.Username, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if err := a.revokeUserTokens(c, user.Uid); err != nil {
		log.Errorf(c, "[administration.UpdateUserPasswordHandler] failed to revoke sessions for user \"%s\", because %s", request.Username, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[administration.UpdateUserPasswordHandler] administrator reset password for user \"%s\"", request.Username)
	return true, nil
}

// DeleteUserHandler removes a member account and revokes their active sessions.
func (a *AdministrationApi) DeleteUserHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.authorize(c); err != nil {
		return nil, err
	}

	user, requestErr := a.getRequestedUser(c, "DeleteUserHandler")

	if requestErr != nil {
		return nil, requestErr
	}

	if err := a.users.DeleteUser(c, user.Username); err != nil {
		log.Errorf(c, "[administration.DeleteUserHandler] failed to delete user \"%s\", because %s", user.Username, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if err := a.revokeUserTokens(c, user.Uid); err != nil {
		log.Errorf(c, "[administration.DeleteUserHandler] failed to revoke sessions for user \"%s\", because %s", user.Username, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[administration.DeleteUserHandler] administrator deleted user \"%s\"", user.Username)
	return true, nil
}

// ClearUserDataHandler deletes a member's financial data but retains the member account.
func (a *AdministrationApi) ClearUserDataHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.authorize(c); err != nil {
		return nil, err
	}

	user, requestErr := a.getRequestedUser(c, "ClearUserDataHandler")

	if requestErr != nil {
		return nil, requestErr
	}

	if err := a.templates.DeleteAllTemplates(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "transaction templates", err)
	}

	if err := a.transactions.DeleteAllTransactions(c, user.Uid, true); err != nil {
		return a.clearUserDataError(c, user.Username, "transactions and accounts", err)
	}

	if err := a.categories.DeleteAllCategories(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "transaction categories", err)
	}

	if err := a.tags.DeleteAllTags(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "transaction tags", err)
	}

	if err := a.tagGroups.DeleteAllTagGroups(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "transaction tag groups", err)
	}

	if err := a.userCustomIcons.DeleteAllCustomIcons(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "custom icons", err)
	}

	if err := a.userCustomExchangeRates.DeleteAllCustomExchangeRates(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "custom exchange rates", err)
	}

	if err := a.insightsExplorers.DeleteAllExplorations(c, user.Uid); err != nil {
		return a.clearUserDataError(c, user.Username, "insight explorations", err)
	}

	if err := a.revokeUserTokens(c, user.Uid); err != nil {
		log.Errorf(c, "[administration.ClearUserDataHandler] failed to revoke sessions for user \"%s\", because %s", user.Username, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[administration.ClearUserDataHandler] administrator cleared data for user \"%s\"", user.Username)
	return true, nil
}

func (a *AdministrationApi) authorize(c *core.WebContext) *errs.Error {
	adminPassword := a.CurrentConfig().AdminPassword
	suppliedPassword := c.GetHeader(adminPasswordHeaderName)

	if adminPassword == "" || suppliedPassword == "" || subtle.ConstantTimeCompare([]byte(suppliedPassword), []byte(adminPassword)) != 1 {
		log.Warnf(c, "[administration.authorize] denied administration request")
		return errs.ErrUnauthorizedAccess
	}

	return nil
}

func (a *AdministrationApi) getRequestedUser(c *core.WebContext, handlerName string) (*models.User, *errs.Error) {
	var request models.AdminUserActionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Warnf(c, "[administration.%s] parse request failed, because %s", handlerName, err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	request.Username = strings.TrimSpace(request.Username)
	user, err := a.users.GetUserByUsername(c, request.Username)

	if err != nil {
		log.Warnf(c, "[administration.%s] failed to get user \"%s\", because %s", handlerName, request.Username, err.Error())
		return nil, errs.ErrUserNotFound
	}

	return user, nil
}

func (a *AdministrationApi) revokeUserTokens(c *core.WebContext, uid int64) error {
	tokens, err := a.tokens.GetAllTokensByUid(c, uid)

	if err != nil {
		return err
	}

	if len(tokens) < 1 {
		return nil
	}

	return a.tokens.DeleteTokens(c, uid, tokens)
}

func (a *AdministrationApi) clearUserDataError(c *core.WebContext, username string, dataType string, err error) (any, *errs.Error) {
	log.Errorf(c, "[administration.ClearUserDataHandler] failed to delete %s for user \"%s\", because %s", dataType, username, err.Error())
	return nil, errs.Or(err, errs.ErrOperationFailed)
}
