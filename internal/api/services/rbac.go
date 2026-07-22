package services

import (
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

// EvaluatePolicy checks explicit ABAC policies first, then falls back to
// role-based rules. Returns true if action is allowed. An error is returned
// if the policy lookup fails (caller must decide; default deny).
func EvaluatePolicy(
	db *gorm.DB,
	subjectType, subjectID, action, resourceType, resourceID string,
) (bool, error) {
	polRepo := repository.NewPolicyRepository(db)

	// Explicit deny wins.
	policies, err := polRepo.FindApplicable(subjectType, subjectID, action, resourceType, resourceID)
	if err != nil {
		return false, err
	}
	for _, p := range policies {
		if p.Effect == "deny" {
			return false, nil
		}
	}
	for _, p := range policies {
		if p.Effect == "allow" {
			return true, nil
		}
	}

	// Fallback: system admin/owner bypass everything.
	if subjectType == "role" {
		if subjectID == models.RoleAdmin || subjectID == models.RoleOwner {
			return true, nil
		}
	}
	// A user subject whose role is admin/owner also bypasses.
	if subjectType == "user" && subjectID != "" {
		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.FindByID(subjectID)
		if err == nil && (user.Role == models.RoleAdmin || user.Role == models.RoleOwner) {
			return true, nil
		}
	}
	return false, nil
}
