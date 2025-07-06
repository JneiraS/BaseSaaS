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
	LastPaymentDate  *time.Time
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
	UpdateLastPaymentDate(memberID uint, date time.Time) error
	GetTotalMembersCount(userID uint) (int64, error)
	GetMembersCountByStatus(userID uint) (map[models.MembershipStatus]int64, error)
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

// UpdateLastPaymentDate met à jour la date du dernier paiement pour un membre.
func (r *GormMemberRepository) UpdateLastPaymentDate(memberID uint, date time.Time) error {
	return r.db.Model(&MemberDB{}).Where("id = ?", memberID).Update("last_payment_date", date).Error
}

// GetTotalMembersCount retourne le nombre total de membres pour un utilisateur donné.
func (r *GormMemberRepository) GetTotalMembersCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&MemberDB{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetMembersCountByStatus retourne le nombre de membres par statut pour un utilisateur donné.
func (r *GormMemberRepository) GetMembersCountByStatus(userID uint) (map[models.MembershipStatus]int64, error) {
	counts := make(map[models.MembershipStatus]int64)
	var results []struct {
		MembershipStatus models.MembershipStatus
		Count            int64
	}

	if err := r.db.Model(&MemberDB{}).Where("user_id = ?", userID).Group("membership_status").Select("membership_status, count(*) as count").Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, res := range results {
		counts[res.MembershipStatus] = res.Count
	}

	// Assurez-vous que tous les statuts possibles sont présents, même avec un count de 0
	for _, status := range []models.MembershipStatus{models.StatusActive, models.StatusInactive, models.StatusPending, models.StatusExpired} {
		if _, ok := counts[status]; !ok {
			counts[status] = 0
		}
	}

	return counts, nil
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
		LastPaymentDate:  m.LastPaymentDate,
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
		LastPaymentDate:  mdb.LastPaymentDate,
	}
}
