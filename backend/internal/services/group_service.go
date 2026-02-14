package services

import (
	"errors"

	"gorm.io/gorm"
	"onechat/internal/models"
)

type GroupService struct {
	db *gorm.DB
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{db: db}
}

func (s *GroupService) CreateGroup(name, description, icon string, createdByID uint, memberIDs []uint) (*models.Group, error) {
	if len(memberIDs) > 256 {
		return nil, errors.New("maximum 256 members allowed")
	}

	// Create group
	group := &models.Group{
		Name:        name,
		Description: description,
		Icon:        icon,
		CreatedByID: createdByID,
	}

	tx := s.db.Begin()
	if err := tx.Create(group).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Add creator as admin
	creatorMember := &models.GroupMember{
		GroupID: group.ID,
		UserID:  createdByID,
		Role:    "admin",
	}
	if err := tx.Create(creatorMember).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Add other members
	for _, memberID := range memberIDs {
		if memberID != createdByID {
			member := &models.GroupMember{
				GroupID: group.ID,
				UserID:  memberID,
				Role:    "member",
			}
			if err := tx.Create(member).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// Create corresponding chat
	chat := &models.Chat{
		Type:    "group",
		GroupID: &group.ID,
	}
	if err := tx.Create(chat).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload with members
	s.db.Preload("Members.User").Preload("CreatedBy").First(group, group.ID)

	return group, nil
}

func (s *GroupService) GetGroup(groupID uint) (*models.Group, error) {
	var group models.Group
	if err := s.db.Preload("Members.User").Preload("CreatedBy").First(&group, groupID).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *GroupService) UpdateGroup(groupID, userID uint, updates map[string]interface{}) (*models.Group, error) {
	// Check if user is admin
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ? AND role = ?", groupID, userID, "admin").
		First(&member).Error; err != nil {
		return nil, errors.New("only admins can update group")
	}

	var group models.Group
	if err := s.db.First(&group, groupID).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&group).Updates(updates).Error; err != nil {
		return nil, err
	}

	s.db.Preload("Members.User").First(&group, groupID)
	return &group, nil
}

func (s *GroupService) DeleteGroup(groupID, userID uint) error {
	// Check if user is admin
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ? AND role = ?", groupID, userID, "admin").
		First(&member).Error; err != nil {
		return errors.New("only admins can delete group")
	}

	tx := s.db.Begin()

	// Delete all members
	if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete group chat
	if err := tx.Where("group_id = ?", groupID).Delete(&models.Chat{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete group
	if err := tx.Delete(&models.Group{}, groupID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *GroupService) AddMember(groupID, userID, newMemberID uint) error {
	// Check member limit
	var count int64
	s.db.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count)
	if count >= 256 {
		return errors.New("group has reached maximum capacity")
	}

	// Check if requester is admin
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ? AND role = ?", groupID, userID, "admin").
		First(&member).Error; err != nil {
		return errors.New("only admins can add members")
	}

	// Check if user already a member
	var existing models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, newMemberID).
		First(&existing).Error; err == nil {
		return errors.New("user is already a member")
	}

	newMember := &models.GroupMember{
		GroupID: groupID,
		UserID:  newMemberID,
		Role:    "member",
	}

	return s.db.Create(newMember).Error
}

func (s *GroupService) RemoveMember(groupID, userID, memberToRemoveID uint) error {
	// Check if requester is admin
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ? AND role = ?", groupID, userID, "admin").
		First(&member).Error; err != nil {
		return errors.New("only admins can remove members")
	}

	// Can't remove yourself if you're the only admin
	if userID == memberToRemoveID {
		var adminCount int64
		s.db.Model(&models.GroupMember{}).
			Where("group_id = ? AND role = ?", groupID, "admin").
			Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("cannot remove the only admin")
		}
	}

	return s.db.Where("group_id = ? AND user_id = ?", groupID, memberToRemoveID).
		Delete(&models.GroupMember{}).Error
}

func (s *GroupService) UpdateMemberRole(groupID, userID, memberID uint, newRole string) error {
	if newRole != "admin" && newRole != "member" {
		return errors.New("invalid role")
	}

	// Check if requester is admin
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ? AND role = ?", groupID, userID, "admin").
		First(&member).Error; err != nil {
		return errors.New("only admins can change roles")
	}

	return s.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, memberID).
		Update("role", newRole).Error
}
