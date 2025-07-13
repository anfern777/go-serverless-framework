package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
)

type UserGroup string
type UserGroupKey string

const USER_GROUP_KEY UserGroupKey = "UserGroup"

const (
	ADMIN UserGroup = "admin"
	USER  UserGroup = "user"
)

var validGroups = []UserGroup{
	ADMIN, USER,
}

func (ug UserGroup) isValid() bool {
	return slices.Contains(validGroups, ug)
}

func AuthorizationMiddleware(next Handler, logger services.Logger, allowedGroups ...UserGroup) Handler {
	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
		jwtClaims := req.RequestContext.Authorizer.JWT.Claims
		group, err := GetUserGroupFromJWTClaims(jwtClaims)
		logger.Log(
			"User Group information retrieved from JWT claims",
			slog.LevelInfo,
			"UserGroup", group,
		)
		if err != nil {
			return common.RequestErrorResponse(http.StatusForbidden, fmt.Sprintf("Failed to get user group %v", err), logger)
		}
		if !slices.Contains(allowedGroups, group) {
			return common.RequestErrorResponse(http.StatusForbidden, "Not authorized to perform operation", logger)
		}
		logger.Log(
			"User is whitelisted",
			slog.LevelInfo,
			"AllowedGroups", allowedGroups[0],
		)

		newCtx := context.WithValue(ctx, USER_GROUP_KEY, group)
		return next(newCtx, req)
	}
}

func GetUserGroupFromJWTClaims(claims map[string]string) (UserGroup, error) {
	userGroupStr, ok := claims["cognito:groups"]
	if !ok {
		return "", fmt.Errorf("user group (cognito:groups claim) not found in JWT claims")
	}

	userGroupStr, found1 := strings.CutPrefix(userGroupStr, "[")
	userGroupStr, found2 := strings.CutSuffix(userGroupStr, "]")

	if !found1 || !found2 {
		return "", fmt.Errorf("failed to inspect user group from JWT group claims: %s", userGroupStr)
	}

	if !UserGroup(userGroupStr).isValid() {
		return "", fmt.Errorf("invalid user group: %s", userGroupStr)
	}
	return UserGroup(userGroupStr), nil
}
