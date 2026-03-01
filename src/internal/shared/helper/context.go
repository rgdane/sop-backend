package helper

import (
	"context"
	"jk-api/pkg/gorm/audit"

	"github.com/gofiber/fiber/v2"
)

// ExtractUserIDFromContext retrieves user ID from Fiber context
func ExtractUserIDFromContext(c *fiber.Ctx) *int64 {
	if userID, ok := c.Locals(string(audit.UserIDKey)).(float64); ok {
		id := int64(userID)
		return &id
	}
	if userID, ok := c.Locals(string(audit.UserIDKey)).(int64); ok {
		return &userID
	}
	if userID, ok := c.Locals(string(audit.UserIDKey)).(int); ok {
		id := int64(userID)
		return &id
	}
	return nil
}

// ExtractIPAddressFromContext retrieves IP address from Fiber context
func ExtractIPAddressFromContext(c *fiber.Ctx) *string {
	ip := c.IP()
	if ip == "" {
		return nil
	}
	return &ip
}

// ExtractUserAgentFromContext retrieves user agent from Fiber context
func ExtractUserAgentFromContext(c *fiber.Ctx) *string {
	ua := c.Get("User-Agent")
	if ua == "" {
		return nil
	}
	return &ua
}

// InjectContextToGorm creates a GORM context with user metadata
func InjectContextToGorm(c *fiber.Ctx) context.Context {
	ctx := context.Background()

	if userID := ExtractUserIDFromContext(c); userID != nil {
		ctx = context.WithValue(ctx, audit.UserIDKey, *userID)
	}

	if ip := ExtractIPAddressFromContext(c); ip != nil {
		ctx = context.WithValue(ctx, audit.IPAddressKey, *ip)
	}

	if ua := ExtractUserAgentFromContext(c); ua != nil {
		ctx = context.WithValue(ctx, audit.UserAgentKey, *ua)
	}

	return ctx
} // GetUserIDFromGormContext retrieves user ID from GORM context
func GetUserIDFromGormContext(ctx context.Context) *int64 {
	if userID, ok := ctx.Value(audit.UserIDKey).(int64); ok {
		return &userID
	}
	return nil
}

// GetIPAddressFromGormContext retrieves IP address from GORM context
func GetIPAddressFromGormContext(ctx context.Context) *string {
	if ip, ok := ctx.Value(audit.IPAddressKey).(string); ok {
		return &ip
	}
	return nil
}

// GetUserAgentFromGormContext retrieves user agent from GORM context
func GetUserAgentFromGormContext(ctx context.Context) *string {
	if ua, ok := ctx.Value(audit.UserAgentKey).(string); ok {
		return &ua
	}
	return nil
}
