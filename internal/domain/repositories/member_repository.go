package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// MemberDB represents the database model for a member, used for GORM persistence.
// It includes GORM's Model for common fields like ID, CreatedAt, UpdatedAt, DeletedAt.
type MemberDB struct {
	gorm.Model
	FirstName        string                // First name of the member
	LastName         string                // Last name of the member
	Email            string                // Email address of the member
	UserID           uint                  // Foreign key linking to the User who owns this member record
	MembershipStatus models.MembershipStatus // Current status of the member's membership
	JoinDate         time.Time             // Date when the member joined
	EndDate          *time.Time            // Optional end date of the membership
	LastPaymentDate  *time.Time            // Optional date of the last payment received from the member
}

// TableName specifies the table name for the MemberDB model in the database.
// This overrides GORM's default naming convention.
func (MemberDB) TableName() string {
	return "members"
}

// MemberRepository defines the interface for member persistence operations.
// It abstracts the underlying database implementation, allowing for different
// data storage mechanisms (e.g., GORM, SQL, NoSQL) to be used interchangeably.
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

// GormMemberRepository is an implementation of MemberRepository that uses GORM
// for interacting with a relational database.
type GormMemberRepository struct {
	db *gorm.DB // GORM database client
}

// NewGormMemberRepository creates a new instance of GormMemberRepository.
// It takes a GORM DB instance as a dependency.
func NewGormMemberRepository(db *gorm.DB) *GormMemberRepository {
	return &GormMemberRepository{db: db}
}

// CreateMember persists a new member to the database.
// It converts the domain model Member to a database-specific MemberDB model
// before saving and then updates the domain model with the generated ID.
func (r *GormMemberRepository) CreateMember(member *models.Member) error {
	memberDB := toMemberDB(member)
	if err := r.db.Create(&memberDB).Error; err != nil {
		return err
	}
	*member = *toMember(memberDB) // Update the original member with DB-generated fields (e.g., ID)
	return nil
}

// FindMemberByID retrieves a member from the database by its ID.
// It returns the member as a domain model or an error if not found.
func (r *GormMemberRepository) FindMemberByID(id uint) (*models.Member, error) {
	var memberDB MemberDB
	result := r.db.First(&memberDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toMember(&memberDB), nil
}

// FindMembersByUserID retrieves all members associated with a specific user ID.
// It queries the database for members where the UserID matches the provided ID.
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

// UpdateMember updates an existing member in the database.
// It converts the domain model to a database model and saves the changes.
func (r *GormMemberRepository) UpdateMember(member *models.Member) error {
	memberDB := toMemberDB(member)
	return r.db.Save(&memberDB).Error
}

// DeleteMember deletes a member from the database by its ID.
func (r *GormMemberRepository) DeleteMember(id uint) error {
	return r.db.Delete(&MemberDB{}, id).Error
}

// UpdateLastPaymentDate updates the last_payment_date field for a specific member.
// It finds the member by ID and updates only the specified field.
func (r *GormMemberRepository) UpdateLastPaymentDate(memberID uint, date time.Time) error {
	return r.db.Model(&MemberDB{}).Where("id = ?", memberID).Update("last_payment_date", date).Error
}

// GetTotalMembersCount returns the total number of members for a given user ID.
// It performs a count query on the members table, filtered by user_id.
func (r *GormMemberRepository) GetTotalMembersCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&MemberDB{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetMembersCountByStatus returns the number of members grouped by their membership status
// for a given user ID. It ensures all possible statuses are included, even if their count is zero.
func (r *GormMemberRepository) GetMembersCountByStatus(userID uint) (map[models.MembershipStatus]int64, error) {
	counts := make(map[models.MembershipStatus]int64)
	var results []struct {
		MembershipStatus models.MembershipStatus
		Count            int64
	}

	// Group members by status and count them.
	if err := r.db.Model(&MemberDB{}).Where("user_id = ?", userID).Group("membership_status").Select("membership_status, count(*) as count").Scan(&results).Error; err != nil {
		return nil, err
	}

	// Populate the map with retrieved counts.
	for _, res := range results {
		counts[res.MembershipStatus] = res.Count
	}

	// Ensure all possible statuses are present in the map, even if their count is 0.
	for _, status := range []models.MembershipStatus{models.StatusActive, models.StatusInactive, models.StatusPending, models.StatusExpired} {
		if _, ok := counts[status]; !ok {
			counts[status] = 0
		}
	}

	return counts, nil
}

// toMemberDB converts a domain Member model to a database-specific MemberDB model.
// This is used before persisting the member to the database.
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

// toMember converts a database-specific MemberDB model back to a domain Member model.
// This is used after retrieving data from the database.
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
