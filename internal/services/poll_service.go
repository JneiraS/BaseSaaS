package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// PollService encapsule la logique métier pour la gestion des sondages.
type PollService struct {
	pollRepo repositories.PollRepository
	voteRepo repositories.VoteRepository
}

// NewPollService crée une nouvelle instance de PollService.
func NewPollService(pollRepo repositories.PollRepository, voteRepo repositories.VoteRepository) *PollService {
	return &PollService{pollRepo: pollRepo, voteRepo: voteRepo}
}

// CreatePoll crée un nouveau sondage avec ses options.
func (s *PollService) CreatePoll(poll *models.Poll) error {
	if err := s.validatePoll(poll); err != nil {
		return err
	}
	return s.pollRepo.CreatePoll(poll)
}

// GetPollByID récupère un sondage par son ID.
func (s *PollService) GetPollByID(id uint) (*models.Poll, error) {
	return s.pollRepo.FindPollByID(id)
}

// GetAllPolls récupère tous les sondages.
func (s *PollService) GetAllPolls() ([]models.Poll, error) {
	return s.pollRepo.FindAllPolls()
}

// GetPollsByUserID récupère tous les sondages créés par un utilisateur.
func (s *PollService) GetPollsByUserID(userID uint) ([]models.Poll, error) {
	return s.pollRepo.FindPollsByUserID(userID)
}

// UpdatePoll met à jour un sondage existant.
func (s *PollService) UpdatePoll(poll *models.Poll) error {
	if err := s.validatePoll(poll); err != nil {
		return err
	}
	return s.pollRepo.UpdatePoll(poll)
}

// DeletePoll supprime un sondage.
func (s *PollService) DeletePoll(id uint) error {
	return s.pollRepo.DeletePoll(id)
}

// Vote enregistre un vote pour une option donnée.
func (s *PollService) Vote(optionID, userID, pollID uint) error {
	// Vérifier si l'utilisateur a déjà voté pour ce sondage
	hasVoted, err := s.voteRepo.HasUserVoted(userID, pollID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification du vote: %w", err)
	}
	if hasVoted {
		return fmt.Errorf("vous avez déjà voté pour ce sondage")
	}

	// Vérifier que l'option appartient bien au sondage
	poll, err := s.pollRepo.FindPollByID(pollID)
	if err != nil {
		return fmt.Errorf("sondage non trouvé: %w", err)
	}

	optionExists := false
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			optionExists = true
			break
		}
	}

	if !optionExists {
		return fmt.Errorf("l'option de vote spécifiée n'appartient pas à ce sondage")
	}

	vote := &models.Vote{
		OptionID: optionID,
		UserID:   userID,
	}
	return s.voteRepo.CreateVote(vote)
}

// GetPollResults récupère les résultats d'un sondage.
func (s *PollService) GetPollResults(pollID uint) (map[uint]int64, error) {
	return s.pollRepo.GetPollResults(pollID)
}

// HasUserVoted vérifie si un utilisateur a déjà voté pour un sondage donné.
func (s *PollService) HasUserVoted(userID, pollID uint) (bool, error) {
	return s.voteRepo.HasUserVoted(userID, pollID)
}

// validatePoll valide les données d'un sondage.
func (s *PollService) validatePoll(poll *models.Poll) error {
	poll.Question = strings.TrimSpace(poll.Question)

	if poll.Question == "" {
		return fmt.Errorf("la question du sondage est requise")
	}
	if len(poll.Options) < 2 {
		return fmt.Errorf("un sondage doit avoir au moins deux options")
	}

	for i, opt := range poll.Options {
		poll.Options[i].Text = strings.TrimSpace(opt.Text)
		if poll.Options[i].Text == "" {
			return fmt.Errorf("le texte de l'option %d est requis", i+1)
		}
	}

	return nil
}
