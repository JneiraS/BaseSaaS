package repositories

import (
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// PollDB représente le modèle de sondage pour la persistance GORM.
type PollDB struct {
	gorm.Model
	Question string
	UserID   uint
	Options  []OptionDB `gorm:"foreignKey:PollID"`
}

// OptionDB représente le modèle d'option pour la persistance GORM.
type OptionDB struct {
	gorm.Model
	Text   string
	PollID uint
	Votes  []VoteDB `gorm:"foreignKey:OptionID"`
}

// VoteDB représente le modèle de vote pour la persistance GORM.
type VoteDB struct {
	gorm.Model
	OptionID uint
	UserID   uint
}

// TableName spécifie le nom de la table pour le modèle PollDB.
func (PollDB) TableName() string {
	return "polls"
}

// TableName spécifie le nom de la table pour le modèle OptionDB.
func (OptionDB) TableName() string {
	return "options"
}

// TableName spécifie le nom de la table pour le modèle VoteDB.
func (VoteDB) TableName() string {
	return "votes"
}

// PollRepository définit l'interface pour les opérations de persistance des sondages.
type PollRepository interface {
	CreatePoll(poll *models.Poll) error
	FindPollByID(id uint) (*models.Poll, error)
	FindPollsByUserID(userID uint) ([]models.Poll, error)
	FindAllPolls() ([]models.Poll, error)
	UpdatePoll(poll *models.Poll) error
	DeletePoll(id uint) error
	GetPollResults(pollID uint) (map[uint]int64, error) // Retourne OptionID -> Count
}

// GormPollRepository est une implémentation de PollRepository utilisant GORM.
type GormPollRepository struct {
	db *gorm.DB
}

// NewGormPollRepository crée une nouvelle instance de GormPollRepository.
func NewGormPollRepository(db *gorm.DB) *GormPollRepository {
	return &GormPollRepository{db: db}
}

// CreatePoll crée un nouveau sondage avec ses options.
func (r *GormPollRepository) CreatePoll(poll *models.Poll) error {
	pollDB := toPollDB(poll)
	if err := r.db.Create(&pollDB).Error; err != nil {
		return err
	}
	*poll = *toPoll(pollDB)
	return nil
}

// FindPollByID recherche un sondage par son ID, avec ses options.
func (r *GormPollRepository) FindPollByID(id uint) (*models.Poll, error) {
	var pollDB PollDB
	result := r.db.Preload("Options").First(&pollDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toPoll(&pollDB), nil
}

// FindPollsByUserID recherche tous les sondages créés par un utilisateur donné.
func (r *GormPollRepository) FindPollsByUserID(userID uint) ([]models.Poll, error) {
	var pollsDB []PollDB
	if err := r.db.Preload("Options").Where("user_id = ?", userID).Find(&pollsDB).Error; err != nil {
		return nil, err
	}
	var polls []models.Poll
	for _, pdb := range pollsDB {
		polls = append(polls, *toPoll(&pdb))
	}
	return polls, nil
}

// FindAllPolls recherche tous les sondages, avec leurs options.
func (r *GormPollRepository) FindAllPolls() ([]models.Poll, error) {
	var pollsDB []PollDB
	if err := r.db.Preload("Options").Find(&pollsDB).Error; err != nil {
		return nil, err
	}
	var polls []models.Poll
	for _, pdb := range pollsDB {
		polls = append(polls, *toPoll(&pdb))
	}
	return polls, nil
}

// UpdatePoll met à jour un sondage existant et ses options.
func (r *GormPollRepository) UpdatePoll(poll *models.Poll) error {
	pollDB := toPollDB(poll)
	// GORM ne met pas à jour les relations imbriquées par défaut avec Save.
	// Il faut gérer les options séparément ou utiliser des transactions.
	return r.db.Save(&pollDB).Error
}

// DeletePoll supprime un sondage par son ID, y compris ses options et votes associés.
func (r *GormPollRepository) DeletePoll(id uint) error {
	// Supprimer les votes associés aux options du sondage
	if err := r.db.Where("option_id IN (SELECT id FROM options WHERE poll_id = ?)", id).Delete(&VoteDB{}).Error; err != nil {
		return err
	}
	// Supprimer les options du sondage
	if err := r.db.Where("poll_id = ?", id).Delete(&OptionDB{}).Error; err != nil {
		return err
	}
	// Supprimer le sondage lui-même
	return r.db.Delete(&PollDB{}, id).Error
}

// GetPollResults retourne les résultats d'un sondage (nombre de votes par option).
func (r *GormPollRepository) GetPollResults(pollID uint) (map[uint]int64, error) {
	results := make(map[uint]int64)
	var rawResults []struct {
		OptionID uint
		Count    int64
	}

	if err := r.db.Model(&VoteDB{}).Select("option_id, count(*) as count").Where("option_id IN (SELECT id FROM options WHERE poll_id = ?)", pollID).Group("option_id").Scan(&rawResults).Error; err != nil {
		return nil, err
	}

	for _, res := range rawResults {
		results[res.OptionID] = res.Count
	}
	return results, nil
}

// toPollDB convertit un modèle de domaine Poll en un modèle de base de données PollDB.
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

// toPoll convertit un modèle de base de données PollDB en un modèle de domaine Poll.
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

// toOptionDB convertit un modèle de domaine Option en un modèle de base de données OptionDB.
func toOptionDB(o *models.Option) *OptionDB {
	return &OptionDB{
		Model:  gorm.Model{ID: o.ID, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt, DeletedAt: o.DeletedAt},
		Text:   o.Text,
		PollID: o.PollID,
	}
}

// toOption convertit un modèle de base de données OptionDB en un modèle de domaine Option.
func toOption(odb *OptionDB) *models.Option {
	return &models.Option{
		Model:  gorm.Model{ID: odb.ID, CreatedAt: odb.CreatedAt, UpdatedAt: odb.UpdatedAt, DeletedAt: odb.DeletedAt},
		Text:   odb.Text,
		PollID: odb.PollID,
	}
}

// VoteRepository définit l'interface pour les opérations de persistance des votes.
type VoteRepository interface {
	CreateVote(vote *models.Vote) error
	HasUserVoted(userID, pollID uint) (bool, error)
	GetVotesByOptionID(optionID uint) ([]models.Vote, error)
}

// GormVoteRepository est une implémentation de VoteRepository utilisant GORM.
type GormVoteRepository struct {
	db *gorm.DB
}

// NewGormVoteRepository crée une nouvelle instance de GormVoteRepository.
func NewGormVoteRepository(db *gorm.DB) *GormVoteRepository {
	return &GormVoteRepository{db: db}
}

// CreateVote crée un nouveau vote.
func (r *GormVoteRepository) CreateVote(vote *models.Vote) error {
	voteDB := toVoteDB(vote)
	if err := r.db.Create(&voteDB).Error; err != nil {
		return err
	}
	*vote = *toVote(voteDB)
	return nil
}

// HasUserVoted vérifie si un utilisateur a déjà voté pour un sondage donné.
func (r *GormVoteRepository) HasUserVoted(userID, pollID uint) (bool, error) {
	var count int64
	// Compte les votes de l'utilisateur pour les options appartenant à ce sondage
	if err := r.db.Model(&VoteDB{}).Where("user_id = ? AND option_id IN (SELECT id FROM options WHERE poll_id = ?)", userID, pollID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetVotesByOptionID récupère tous les votes pour une option donnée.
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

// toVoteDB convertit un modèle de domaine Vote en un modèle de base de données VoteDB.
func toVoteDB(v *models.Vote) *VoteDB {
	return &VoteDB{
		Model:    gorm.Model{ID: v.ID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, DeletedAt: v.DeletedAt},
		OptionID: v.OptionID,
		UserID:   v.UserID,
	}
}

// toVote convertit un modèle de base de données VoteDB en un modèle de domaine Vote.
func toVote(vdb *VoteDB) *models.Vote {
	return &models.Vote{
		Model:    gorm.Model{ID: vdb.ID, CreatedAt: vdb.CreatedAt, UpdatedAt: vdb.UpdatedAt, DeletedAt: vdb.DeletedAt},
		OptionID: vdb.OptionID,
		UserID:   vdb.UserID,
	}
}
