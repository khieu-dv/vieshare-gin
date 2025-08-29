package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Notifications handlers for PocketBase compatibility

func (p *PocketBaseController) listNotifications(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var notifications []models.Notification
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM notifications"
	countQuery := "SELECT COUNT(*) FROM notifications"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count notifications"})
		return
	}
	
	// Add sorting
	sortClause := buildSortClause(sort)
	if sortClause == "" {
		sortClause = "ORDER BY created DESC"
	}
	baseQuery += " " + sortClause
	
	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, perPage, offset)
	
	// Execute query
	rows, err := dbMap.Db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var notification models.Notification
		var user sql.NullString
		
		err := rows.Scan(&notification.ID, &notification.Created, &notification.Updated, &notification.CollectionID, 
			&notification.CollectionName, &notification.Email, &notification.Token, &user,
			&notification.Communication, &notification.Newsletter, &notification.Marketing)
		if err != nil {
			continue
		}
		
		if user.Valid {
			notification.User = user.String
		}
		notification.CollectionID = "notifications"
		notification.CollectionName = "notifications"
		notifications = append(notifications, notification)
	}
	
	// Handle expand relations
	if expand != "" {
		notifications = p.expandNotificationRelations(notifications, expand)
	}
	
	// Calculate pagination info
	totalPages := int(totalItems) / perPage
	if int(totalItems)%perPage > 0 {
		totalPages++
	}
	
	response := models.PBListResponse{
		Page:       page,
		PerPage:    perPage,
		TotalItems: int(totalItems),
		TotalPages: totalPages,
		Items:      notifications,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getNotification(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var notification models.Notification
	var user sql.NullString
	
	query := "SELECT * FROM notifications WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&notification.ID, &notification.Created, &notification.Updated, 
		&notification.CollectionID, &notification.CollectionName, &notification.Email, &notification.Token, 
		&user, &notification.Communication, &notification.Newsletter, &notification.Marketing)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification"})
		}
		return
	}
	
	if user.Valid {
		notification.User = user.String
	}
	notification.CollectionID = "notifications"
	notification.CollectionName = "notifications"
	
	// Handle expand relations
	if expand != "" {
		notifications := []models.Notification{notification}
		notifications = p.expandNotificationRelations(notifications, expand)
		if len(notifications) > 0 {
			notification = notifications[0]
		}
	}
	
	c.JSON(http.StatusOK, notification)
}

func (p *PocketBaseController) createNotification(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var notification models.Notification
	setBaseRecord(&notification.BaseRecord, "notifications")
	
	// Map data to notification struct
	if email, ok := data["email"].(string); ok {
		notification.Email = email
	}
	if token, ok := data["token"].(string); ok {
		notification.Token = token
	}
	if user, ok := data["user"].(string); ok && user != "" {
		notification.User = user
	}
	if communication, ok := data["communication"].(bool); ok {
		notification.Communication = communication
	}
	if newsletter, ok := data["newsletter"].(bool); ok {
		notification.Newsletter = newsletter
	}
	if marketing, ok := data["marketing"].(bool); ok {
		notification.Marketing = marketing
	}
	
	// Insert into database
	query := `INSERT INTO notifications (id, created, updated, collection_id, collection_name, email, token, 
		user, communication, newsletter, marketing) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	var userValue interface{}
	if notification.User != "" {
		userValue = notification.User
	}
		
	_, err := dbMap.Db.Exec(query, notification.ID, notification.Created, notification.Updated, notification.CollectionID, 
		notification.CollectionName, notification.Email, notification.Token, userValue, 
		notification.Communication, notification.Newsletter, notification.Marketing)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}
	
	c.JSON(http.StatusOK, notification)
}

func (p *PocketBaseController) updateNotification(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if notification exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM notifications WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if email, ok := data["email"].(string); ok {
		setParts = append(setParts, "email = ?")
		args = append(args, email)
	}
	if token, ok := data["token"].(string); ok {
		setParts = append(setParts, "token = ?")
		args = append(args, token)
	}
	if user, ok := data["user"].(string); ok {
		if user == "" {
			setParts = append(setParts, "user = NULL")
		} else {
			setParts = append(setParts, "user = ?")
			args = append(args, user)
		}
	}
	if communication, ok := data["communication"].(bool); ok {
		setParts = append(setParts, "communication = ?")
		args = append(args, communication)
	}
	if newsletter, ok := data["newsletter"].(bool); ok {
		setParts = append(setParts, "newsletter = ?")
		args = append(args, newsletter)
	}
	if marketing, ok := data["marketing"].(bool); ok {
		setParts = append(setParts, "marketing = ?")
		args = append(args, marketing)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE notifications SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}
	
	// Return updated notification
	p.getNotification(c, id, "")
}

func (p *PocketBaseController) deleteNotification(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM notifications WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandNotificationRelations(notifications []models.Notification, expand string) []models.Notification {
	if expand == "" {
		return notifications
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, notification := range notifications {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "user":
				if notification.User != "" {
					var user models.User
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, email, email_visibility, username, name, avatar, verified FROM users WHERE id = ?", 
						notification.User).Scan(&user.ID, &user.Created, &user.Updated, &user.CollectionID, 
						&user.CollectionName, &user.Email, &user.EmailVisibility, &user.Username, 
						&user.Name, &user.Avatar, &user.Verified)
					if err == nil {
						user.CollectionID = "users"
						user.CollectionName = "users"
						expandData["user"] = user
					}
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		notifications[i] = notification
	}
	
	return notifications
}