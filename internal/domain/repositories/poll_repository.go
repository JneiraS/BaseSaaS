package repositories

import (
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// PollDB represents the database model for a poll, used for GORM persistence.
// It includes GORM's Model for common fields and has a one-to-many relationship with OptionDB.
type PollDB struct {
	gorm.Model
	Question string     // The question of the poll.
	UserID   uint       // The ID of the user who created the poll.
	Options  []OptionDB `gorm:"foreignKey:PollID"` // Associated options for this poll.
}

// OptionDB represents the database model for a poll option, used for GORM persistence.
// It includes GORM's Model for common fields and has a one-to-many relationship with VoteDB.
type OptionDB struct {
	gorm.Model
	Text   string   // The text of the option.
	PollID uint     // The ID of the poll this option belongs to.
	Votes  []VoteDB `gorm:"foreignKey:OptionID"` // Associated votes for this option.
}

// VoteDB represents the database model for a vote, used for GORM persistence.
// It includes GORM's Model for common fields.
type VoteDB struct {
	gorm.Model
	OptionID uint // The ID of the option that was voted for.
	UserID   uint // The ID of the user who cast the vote.
}

// TableName specifies the table name for the PollDB model.
func (PollDB) TableName() string {
	return "polls"
}

// TableName specifies the table name for the OptionDB model.
func (OptionDB) TableName() string {
	return "options"
}

// TableName specifies the table name for the VoteDB model.
func (VoteDB) TableName() string {
	return "votes"
}

// PollRepository defines the interface for poll persistence operations.
// It abstracts the underlying database implementation for polls and their options.
type PollRepository interface {
	CreatePoll(poll *models.Poll) error
	FindPollByID(id uint) (*models.Poll, error)
	FindPollsByUserID(userID uint) ([]models.Poll, error)
	FindAllPolls() ([]models.Poll, error)
	UpdatePoll(poll *models.Poll) error
	DeletePoll(id uint) error
	GetPollResults(pollID uint) (map[uint]int64, error) // Returns OptionID -> Count of votes.
}

// GormPollRepository is an implementation of PollRepository that uses GORM.
type GormPollRepository struct {
	db *gorm.DB // GORM database client.
}

// NewGormPollRepository creates a new instance of GormPollRepository.
func NewGormPollRepository(db *gorm.DB) *GormPollRepository {
	return &GormPollRepository{db: db}
}

// CreatePoll persists a new poll and its associated options to the database.
func (r *GormPollRepository) CreatePoll(poll *models.Poll) error {
	pollDB := toPollDB(poll)
	if err := r.db.Create(&pollDB).Error; err != nil {
		return err
	}
	*poll = *toPoll(pollDB) // Update the original poll with DB-generated fields (e.g., IDs for poll and options).
	return nil
}

// FindPollByID retrieves a poll by its ID, eagerly loading its options.
func (r *GormPollRepository) FindPollByID(id uint) (*models.Poll, error) {
	var pollDB PollDB
	// Preload "Options" to fetch associated options in a single query.
	result := r.db.Preload("Options").First(&pollDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toPoll(&pollDB), nil
}

// FindPollsByUserID retrieves all polls created by a specific user, eagerly loading their options.
func (r *GormPollRepository) FindPollsByUserID(userID uint) ([]models.Poll, error) {
	var pollsDB []PollDB
	// Preload "Options" and filter by UserID.
	if err := r.db.Preload("Options").Where("user_id = ?", userID).Find(&pollsDB).Error; err != nil {
		return nil, err
	}
	var polls []models.Poll
	for _, pdb := range pollsDB {
		polls = append(polls, *toPoll(&pdb))
	}
	return polls, nil
}

// FindAllPolls retrieves all polls, eagerly loading their options.
func (r *GormPollRepository) FindAllPolls() ([]models.Poll, error) {
	var pollsDB []PollDB
	// Preload "Options" to fetch all polls with their associated options.
	if err := r.db.Preload("Options").Find(&pollsDB).Error; err != nil {
		return nil, err
	}
	var polls []models.Poll
	for _, pdb := range pollsDB {
		polls = append(polls, *toPoll(&pdb))
	}
	return polls, nil
}

// UpdatePoll updates an existing poll and its options.
// Note: GORM's Save method does not automatically update nested associations.
// For complex updates involving options, manual handling or transactions might be needed.
func (r *GormPollRepository) UpdatePoll(poll *models.Poll) error {
	pollDB := toPollDB(poll)
	// GORM does not update nested relations by default with Save.
	// Options need to be managed separately or within a transaction.
	return r.db.Save(&pollDB).Error
}

// DeletePoll deletes a poll by its ID, including its associated options and votes.
// It performs cascading deletes for related records.
func (r *GormPollRepository) DeletePoll(id uint) error {
	// Delete votes associated with the poll's options.
	if err := r.db.Where("option_id IN (SELECT id FROM options WHERE poll_id = ?)", id).Delete(&VoteDB{}).Error; err != nil {
		return err
	}
	// Delete options belonging to the poll.
	if err := r.db.Where("poll_id = ?", id).Delete(&OptionDB{}).Error; err != nil {
		return err
	}
	// Delete the poll itself.
	return r.db.Delete(&PollDB{}, id).Error
}

// GetPollResults returns the vote counts for each option of a given poll.
// The result is a map where keys are OptionIDs and values are vote counts.
func (r *GormPollRepository) GetPollResults(pollID uint) (map[uint]int64, error) {
	results := make(map[uint]int64)
	var rawResults []struct {
		OptionID uint
		Count    int64
	}

	// Query to count votes for options belonging to the specified poll.
	if err := r.db.Model(&VoteDB{}).Select("option_id, count(*) as count").Where("option_id IN (SELECT id FROM options WHERE poll_id = ?)", pollID).Group("option_id").Scan(&rawResults).Error; err != nil {
		return nil, err
	}

	// Populate the results map.
	for _, res := range rawResults {
		results[res.OptionID] = res.Count
	}
	return results, nil
}

// toPollDB converts a domain Poll model to a database-specific PollDB model.
// It also converts associated Option models to OptionDB models.
func toPollDB(p *models.Poll) *PollDB {
	pollDB := &PollDB{
		Model:    gorm.Model{ID: p.ID, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, DeletedAt: p.DeletedAt},
		Question: p.Question,
		UserID:   p.UserID,
	}
	for _, opt := range p.Options {
		pollDB.Options = append(pollDB.Options, *toOptionDB(&opt))
	}
	return pollDB
}

// toPoll converts a database-specific PollDB model back to a domain Poll model.
// It also converts associated OptionDB models to Option models.
func toPoll(pdb *PollDB) *models.Poll {
	poll := &models.Poll{
		Model:    gorm.Model{ID: pdb.ID, CreatedAt: pdb.CreatedAt, UpdatedAt: pdb.UpdatedAt, DeletedAt: pdb.DeletedAt},
		Question: pdb.Question,
		UserID:   pdb.UserID,
	}
	for _, optdb := range pdb.Options {
		poll.Options = append(poll.Options, *toOption(&optdb))
	}
	return poll
}

// toOptionDB converts a domain Option model to a database-specific OptionDB model.
func toOptionDB(o *models.Option) *OptionDB {
	return &OptionDB{
		Model:  gorm.Model{ID: o.ID, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt, DeletedAt: o.DeletedAt},
		Text:   o.Text,
		PollID: o.PollID,
	}
}

// toOption converts a database-specific OptionDB model back to a domain Option model.
func toOption(odb *OptionDB) *models.Option {
	return &models.Option{
		Model:  gorm.Model{ID: odb.ID, CreatedAt: odb.CreatedAt, UpdatedAt: odb.UpdatedAt, DeletedAt: odb.DeletedAt},
		Text:   odb.Text,
		PollID: odb.PollID,
	}
}

// VoteRepository defines the interface for vote persistence operations.
// It abstracts the underlying database implementation for votes.
type VoteRepository interface {
	CreateVote(vote *models.Vote) error
	HasUserVoted(userID, pollID uint) (bool, error)
	GetVotesByOptionID(optionID uint) ([]models.Vote, error)
}

// GormVoteRepository is an implementation of VoteRepository that uses GORM.
type GormVoteRepository struct {
	db *gorm.DB // GORM database client.
}

// NewGormVoteRepository creates a new instance of GormVoteRepository.
func NewGormVoteRepository(db *gorm.DB) *GormVoteRepository {
	return &GormVoteRepository{db: db}
}

// CreateVote persists a new vote to the database.
func (r *GormVoteRepository) CreateVote(vote *models.Vote) error {
	voteDB := toVoteDB(vote)
	if err := r.db.Create(&voteDB).Error; err != nil {
		return err
	}
	*vote = *toVote(voteDB) // Update the original vote with DB-generated fields.
	return nil
}

// HasUserVoted checks if a user has already voted in a given poll.
// It queries for votes cast by the user for any option belonging to the specified poll.
func (r *GormVoteRepository) HasUserVoted(userID, pollID uint) (bool, error) {
	var count int64
	// Count votes by the user for options within the specified poll.
	if err := r.db.Model(&VoteDB{}).Where("user_id = ? AND option_id IN (SELECT id FROM options WHERE poll_id = ?)", userID, pollID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetVotesByOptionID retrieves all votes for a specific option.
func (r *GormVoteRepository) GetVotesByOptionID(optionID uint) ([]models.Vote, error) {
	var votesDB []VoteDB
	if err := r.db.Where("option_id = ?", optionID).Find(&votesDB).Error; err != nil {
		return nil, err
	}
	var votes []models.Vote
	for _, vdb := range votesDB {
		votes = append(votes, *toVote(&vdb))
	}
	return votes, nil
}

// toVoteDB converts a domain Vote model to a database-specific VoteDB model.
func toVoteDB(v *models.Vote) *VoteDB {
	return &VoteDB{
		Model:    gorm.Model{ID: v.ID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, DeletedAt: v.DeletedAt},
		OptionID: v.OptionID,
		UserID:   v.UserID,
	}
}

// toVote converts a database-specific VoteDB model back to a domain Vote model.
func toVote(vdb *VoteDB) *models.Vote {
	return &models.Vote{
		Model:    gorm.Model{ID: vdb.ID, CreatedAt: vdb.CreatedAt, UpdatedAt: vdb.UpdatedAt, DeletedAt: vdb.DeletedAt},
		OptionID: vdb.OptionID,
		UserID:   vdb.UserID,
	}
}
