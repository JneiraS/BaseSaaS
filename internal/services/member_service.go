package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// MemberService encapsule la logique métier pour la gestion des membres.
type MemberService struct {
	memberRepo repositories.MemberRepository
}

// NewMemberService crée une nouvelle instance de MemberService.
func NewMemberService(memberRepo repositories.MemberRepository) *MemberService {
	return &MemberService{memberRepo: memberRepo}
}

// CreateMember gère la création d'un nouveau membre.
func (s *MemberService) CreateMember(member *models.Member) error {
	if err := s.validateMember(member); err != nil {
		return err
	}
	return s.memberRepo.CreateMember(member)
}

// GetMemberByID récupère un membre par son ID.
func (s *MemberService) GetMemberByID(id uint) (*models.Member, error) {
	return s.memberRepo.FindMemberByID(id)
}

// GetMembersByUserID récupère tous les membres d'un utilisateur.
func (s *MemberService) GetMembersByUserID(userID uint) ([]models.Member, error) {
	return s.memberRepo.FindMembersByUserID(userID)
}

// UpdateMember gère la mise à jour d'un membre.
func (s *MemberService) UpdateMember(member *models.Member) error {
	if err := s.validateMember(member); err != nil {
		return err
	}
	return s.memberRepo.UpdateMember(member)
}

// DeleteMember gère la suppression d'un membre.
func (s *MemberService) DeleteMember(id uint) error {
	return s.memberRepo.DeleteMember(id)
}

// validateMember valide les données d'un membre.
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
	// Idéalement, ajouter une validation d'email plus robuste ici.

	return nil
}
