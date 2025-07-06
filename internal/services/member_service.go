package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// MemberService encapsulates the business logic for managing members.
// It interacts with the MemberRepository to perform CRUD operations and other member-related tasks.
type MemberService struct {
	memberRepo repositories.MemberRepository
}

// NewMemberService creates a new instance of MemberService.
// It takes a MemberRepository as a dependency, adhering to the dependency inversion principle.
func NewMemberService(memberRepo repositories.MemberRepository) *MemberService {
	return &MemberService{memberRepo: memberRepo}
}

// CreateMember handles the creation of a new member.
// It performs validation on the member data before persisting it via the repository.
func (s *MemberService) CreateMember(member *models.Member) error {
	if err := s.validateMember(member); err != nil {
		return err
	}
	return s.memberRepo.CreateMember(member)
}

// GetMemberByID retrieves a member by its unique identifier.
func (s *MemberService) GetMemberByID(id uint) (*models.Member, error) {
	return s.memberRepo.FindMemberByID(id)
}

// GetMembersByUserID retrieves all members associated with a specific user ID.
func (s *MemberService) GetMembersByUserID(userID uint) ([]models.Member, error) {
	return s.memberRepo.FindMembersByUserID(userID)
}

// UpdateMember handles the update of an existing member.
// It performs validation on the updated member data before persisting the changes.
func (s *MemberService) UpdateMember(member *models.Member) error {
	if err := s.validateMember(member); err != nil {
		return err
	}
	return s.memberRepo.UpdateMember(member)
}

// DeleteMember handles the deletion of a member by its unique identifier.
func (s *MemberService) DeleteMember(id uint) error {
	return s.memberRepo.DeleteMember(id)
}

// MarkPaymentReceived updates the last payment date for a member.
// It calls the repository to update the specific field.
func (s *MemberService) MarkPaymentReceived(memberID uint, paymentDate time.Time) error {
	return s.memberRepo.UpdateLastPaymentDate(memberID, paymentDate)
}

// GetTotalMembersCount returns the total number of members for a given user ID.
// It delegates the call to the underlying repository.
func (s *MemberService) GetTotalMembersCount(userID uint) (int64, error) {
	return s.memberRepo.GetTotalMembersCount(userID)
}

// GetMembersCountByStatus returns the count of members grouped by their membership status
// for a given user ID. It delegates the call to the underlying repository.
func (s *MemberService) GetMembersCountByStatus(userID uint) (map[models.MembershipStatus]int64, error) {
	return s.memberRepo.GetMembersCountByStatus(userID)
}

// validateMember performs business logic validation on a Member model.
// It checks for required fields and can be extended for more complex validations (e.g., email format).
func (s *MemberService) validateMember(member *models.Member) error {
	member.FirstName = strings.TrimSpace(member.FirstName)
	member.LastName = strings.TrimSpace(member.LastName)
	member.Email = strings.TrimSpace(member.Email)

	if member.FirstName == "" {
		return fmt.Errorf("le prénom est requis")
	}
	if member.LastName == "" {
		return fmt.Errorf("le nom de famille est requis")
	}
	if member.Email == "" {
		return fmt.Errorf("l'email est requis")
	}
	// Ideally, add more robust email validation here.

	return nil
}
