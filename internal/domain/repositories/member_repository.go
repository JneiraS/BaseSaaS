package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// MemberDB représente le modèle de membre pour la persistance GORM.
type MemberDB struct {
	gorm.Model
	FirstName        string
	LastName         string
	Email            string
	UserID           uint
	MembershipStatus models.MembershipStatus
	JoinDate         time.Time
	EndDate          *time.Time
}

// TableName spécifie le nom de la table pour le modèle MemberDB.
func (MemberDB) TableName() string {
	return "members"
}

// MemberRepository définit l'interface pour les opérations de persistance des membres.
type MemberRepository interface {
	CreateMember(member *models.Member) error
	FindMemberByID(id uint) (*models.Member, error)
	FindMembersByUserID(userID uint) ([]models.Member, error)
	UpdateMember(member *models.Member) error
	DeleteMember(id uint) error
}

// GormMemberRepository est une implémentation de MemberRepository utilisant GORM.
type GormMemberRepository struct {
	db *gorm.DB
}

// NewGormMemberRepository crée une nouvelle instance de GormMemberRepository.
func NewGormMemberRepository(db *gorm.DB) *GormMemberRepository {
	return &GormMemberRepository{db: db}
}

// CreateMember crée un nouveau membre.
func (r *GormMemberRepository) CreateMember(member *models.Member) error {
	memberDB := toMemberDB(member)
	if err := r.db.Create(&memberDB).Error; err != nil {
		return err
	}
	*member = *toMember(memberDB)
	return nil
}

// FindMemberByID recherche un membre par son ID.
func (r *GormMemberRepository) FindMemberByID(id uint) (*models.Member, error) {
	var memberDB MemberDB
	result := r.db.First(&memberDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toMember(&memberDB), nil
}

// FindMembersByUserID recherche tous les membres pour un utilisateur donné.
func (r *GormMemberRepository) FindMembersByUserID(userID uint) ([]models.Member, error) {
	var membersDB []MemberDB
	if err := r.db.Where("user_id = ?", userID).Find(&membersDB).Error; err != nil {
		return nil, err
	}
	var members []models.Member
	for _, mdb := range membersDB {
		members = append(members, *toMember(&mdb))
	}
	return members, nil
}

// UpdateMember met à jour un membre existant.
func (r *GormMemberRepository) UpdateMember(member *models.Member) error {
	memberDB := toMemberDB(member)
	return r.db.Save(&memberDB).Error
}

// DeleteMember supprime un membre par son ID.
func (r *GormMemberRepository) DeleteMember(id uint) error {
	return r.db.Delete(&MemberDB{}, id).Error
}

// toMemberDB convertit un modèle de domaine Member en un modèle de base de données MemberDB.
func toMemberDB(m *models.Member) *MemberDB {
	return &MemberDB{
		Model:            gorm.Model{ID: m.ID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, DeletedAt: m.DeletedAt},
		FirstName:        m.FirstName,
		LastName:         m.LastName,
		Email:            m.Email,
		UserID:           m.UserID,
		MembershipStatus: m.MembershipStatus,
		JoinDate:         m.JoinDate,
		EndDate:          m.EndDate,
	}
}

// toMember convertit un modèle de base de données MemberDB en un modèle de domaine Member.
func toMember(mdb *MemberDB) *models.Member {
	return &models.Member{
		Model:            gorm.Model{ID: mdb.ID, CreatedAt: mdb.CreatedAt, UpdatedAt: mdb.UpdatedAt, DeletedAt: mdb.DeletedAt},
		FirstName:        mdb.FirstName,
		LastName:         mdb.LastName,
		Email:            mdb.Email,
		UserID:           mdb.UserID,
		MembershipStatus: mdb.MembershipStatus,
		JoinDate:         mdb.JoinDate,
		EndDate:          mdb.EndDate,
	}
}
