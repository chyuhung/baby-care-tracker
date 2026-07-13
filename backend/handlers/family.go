package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMyFamily 获取我的家庭信息
func GetMyFamily(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var familyID int64
	database.DB.QueryRow("SELECT family_id FROM users WHERE id = ?", userID).Scan(&familyID)
	if familyID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未加入家庭"})
		return
	}

	var family models.Family
	err := database.DB.QueryRow(
		"SELECT id, invite_code, created_at FROM families WHERE id = ?",
		familyID,
	).Scan(&family.ID, &family.InviteCode, &family.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "家庭不存在"})
		return
	}

	// 查询家庭成员
	rows, err := database.DB.Query(
		"SELECT id, username FROM users WHERE family_id = ? ORDER BY id",
		familyID,
	)
	var members []models.FamilyMember
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m models.FamilyMember
			if rows.Scan(&m.UserID, &m.Username) == nil {
				m.ID = m.UserID
				members = append(members, m)
			}
		}
	}
	if members == nil {
		members = []models.FamilyMember{}
	}

	c.JSON(http.StatusOK, gin.H{
		"family":  family,
		"members": members,
	})
}

// JoinFamily 通过邀请码加入家庭（自动切换到新家庭）
func JoinFamily(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.JoinFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入邀请码"})
		return
	}

	var familyID int64
	err := database.DB.QueryRow(
		"SELECT id FROM families WHERE invite_code = ?",
		req.InviteCode,
	).Scan(&familyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请码无效"})
		return
	}

	database.DB.Exec("UPDATE users SET family_id = ? WHERE id = ?", familyID, userID)

	c.JSON(http.StatusOK, gin.H{"message": "加入成功", "family_id": familyID})
}

// LeaveFamily 退出家庭
func LeaveFamily(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 创建一个新家庭
	code := generateInviteCode()
	res, err := database.DB.Exec("INSERT INTO families (invite_code) VALUES (?)", code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	newFamilyID, _ := res.LastInsertId()
	database.DB.Exec("UPDATE users SET family_id = ? WHERE id = ?", newFamilyID, userID)

	c.JSON(http.StatusOK, gin.H{"message": "已退出"})
}

// RegenerateInviteCode 重新生成邀请码
func RegenerateInviteCode(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var familyID int64
	database.DB.QueryRow("SELECT family_id FROM users WHERE id = ?", userID).Scan(&familyID)
	if familyID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未加入家庭"})
		return
	}

	code := generateInviteCode()
	database.DB.Exec("UPDATE families SET invite_code = ? WHERE id = ?", code, familyID)

	c.JSON(http.StatusOK, gin.H{"invite_code": code})
}
