package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// PollService encapsulates the business logic for managing polls.
// It interacts with PollRepository and VoteRepository to perform poll and vote-related operations.
type PollService struct {
	pollRepo repositories.PollRepository
	voteRepo repositories.VoteRepository
}

// NewPollService creates a new instance of PollService.
// It takes PollRepository and VoteRepository as dependencies, adhering to the dependency inversion principle.
func NewPollService(pollRepo repositories.PollRepository, voteRepo repositories.VoteRepository) *PollService {
	return &PollService{pollRepo: pollRepo, voteRepo: voteRepo}
}

// CreatePoll handles the creation of a new poll with its options.
// It performs validation on the poll data before persisting it via the repository.
func (s *PollService) CreatePoll(poll *models.Poll) error {
	if err := s.validatePoll(poll); err != nil {
		return err
	}
	return s.pollRepo.CreatePoll(poll)
}

// GetPollByID retrieves a poll by its unique identifier.
func (s *PollService) GetPollByID(id uint) (*models.Poll, error) {
	return s.pollRepo.FindPollByID(id)
}

// GetAllPolls retrieves all polls available in the system.
func (s *PollService) GetAllPolls() ([]models.Poll, error) {
	return s.pollRepo.FindAllPolls()
}

// GetPollsByUserID retrieves all polls created by a specific user.
func (s *PollService) GetPollsByUserID(userID uint) ([]models.Poll, error) {
	return s.pollRepo.FindPollsByUserID(userID)
}

// UpdatePoll handles the update of an existing poll.
// It performs validation on the updated poll data before persisting the changes.
func (s *PollService) UpdatePoll(poll *models.Poll) error {
	if err := s.validatePoll(poll); err != nil {
		return err
	}
	return s.pollRepo.UpdatePoll(poll)
}

// DeletePoll handles the deletion of a poll by its unique identifier.
// This operation typically cascades to delete associated options and votes.
func (s *PollService) DeletePoll(id uint) error {
	return s.pollRepo.DeletePoll(id)
}

// Vote records a user's vote for a given poll option.
// It first checks if the user has already voted in the poll and validates the option.
func (s *PollService) Vote(optionID, userID, pollID uint) error {
	// Check if the user has already voted for this poll.
	hasVoted, err := s.voteRepo.HasUserVoted(userID, pollID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification du vote: %w", err)
	}
	if hasVoted {
		return fmt.Errorf("vous avez déjà voté pour ce sondage")
	}

	// Verify that the option belongs to the specified poll.
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

	// Create and persist the new vote.
	vote := &models.Vote{
		OptionID: optionID,
		UserID:   userID,
	}
	return s.voteRepo.CreateVote(vote)
}

// GetPollResults retrieves the results of a poll (vote counts per option).
func (s *PollService) GetPollResults(pollID uint) (map[uint]int64, error) {
	return s.pollRepo.GetPollResults(pollID)
}

// HasUserVoted checks if a user has already voted in a given poll.
// It delegates the check to the underlying vote repository.
func (s *PollService) HasUserVoted(userID, pollID uint) (bool, error) {
	return s.voteRepo.HasUserVoted(userID, pollID)
}

// validatePoll performs business logic validation on a Poll model.
// It checks for a non-empty question and at least two options, and validates each option's text.
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
