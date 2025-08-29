package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Users handlers for PocketBase compatibility

func (p *PocketBaseController) listUsers(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var users []models.User
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM users"
	countQuery := "SELECT COUNT(*) FROM users"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var user models.User
		
		err := rows.Scan(&user.ID, &user.Created, &user.Updated, &user.CollectionID, &user.CollectionName,
			&user.Email, &user.EmailVisibility, &user.Username, &user.Name, &user.Avatar, &user.Verified)
		if err != nil {
			continue
		}
		
		user.CollectionID = "users"
		user.CollectionName = "users"
		users = append(users, user)
	}
	
	// Handle expand relations
	if expand != "" {
		users = p.expandUserRelations(users, expand)
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
		Items:      users,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getUser(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var user models.User
	
	query := "SELECT * FROM users WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&user.ID, &user.Created, &user.Updated, 
		&user.CollectionID, &user.CollectionName, &user.Email, &user.EmailVisibility, 
		&user.Username, &user.Name, &user.Avatar, &user.Verified)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	user.CollectionID = "users"
	user.CollectionName = "users"
	
	// Handle expand relations
	if expand != "" {
		users := []models.User{user}
		users = p.expandUserRelations(users, expand)
		if len(users) > 0 {
			user = users[0]
		}
	}
	
	c.JSON(http.StatusOK, user)
}

func (p *PocketBaseController) createUser(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var user models.User
	setBaseRecord(&user.BaseRecord, "users")
	
	// Map data to user struct
	if email, ok := data["email"].(string); ok {
		user.Email = email
	}
	if emailVisibility, ok := data["emailVisibility"].(bool); ok {
		user.EmailVisibility = emailVisibility
	}
	if username, ok := data["username"].(string); ok {
		user.Username = username
	}
	if name, ok := data["name"].(string); ok {
		user.Name = name
	}
	if avatar, ok := data["avatar"].(string); ok {
		user.Avatar = avatar
	}
	if verified, ok := data["verified"].(bool); ok {
		user.Verified = verified
	}
	
	// Insert into database
	query := `INSERT INTO users (id, created, updated, collection_id, collection_name, email, email_visibility, 
		username, name, avatar, verified) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err := dbMap.Db.Exec(query, user.ID, user.Created, user.Updated, user.CollectionID, 
		user.CollectionName, user.Email, user.EmailVisibility, user.Username, user.Name, user.Avatar, user.Verified)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	
	c.JSON(http.StatusOK, user)
}

func (p *PocketBaseController) updateUser(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if user exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if email, ok := data["email"].(string); ok {
		setParts = append(setParts, "email = ?")
		args = append(args, email)
	}
	if emailVisibility, ok := data["emailVisibility"].(bool); ok {
		setParts = append(setParts, "email_visibility = ?")
		args = append(args, emailVisibility)
	}
	if username, ok := data["username"].(string); ok {
		setParts = append(setParts, "username = ?")
		args = append(args, username)
	}
	if name, ok := data["name"].(string); ok {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if avatar, ok := data["avatar"].(string); ok {
		setParts = append(setParts, "avatar = ?")
		args = append(args, avatar)
	}
	if verified, ok := data["verified"].(bool); ok {
		setParts = append(setParts, "verified = ?")
		args = append(args, verified)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE users SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	// Return updated user
	p.getUser(c, id, "")
}

func (p *PocketBaseController) deleteUser(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandUserRelations(users []models.User, expand string) []models.User {
	if expand == "" {
		return users
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, user := range users {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "stores":
				var stores []models.Store
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, name, slug, description, user, plan, plan_ends_at, cancel_plan_at_end, product_limit, tag_limit, variant_limit, active FROM stores WHERE user = ?", 
					user.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var store models.Store
						var planEndsAt sql.NullTime
						
						err := rows.Scan(&store.ID, &store.Created, &store.Updated, &store.CollectionID, 
							&store.CollectionName, &store.Name, &store.Slug, &store.Description, &store.User,
							&store.Plan, &planEndsAt, &store.CancelPlanAtEnd, &store.ProductLimit,
							&store.TagLimit, &store.VariantLimit, &store.Active)
						if err == nil {
							if planEndsAt.Valid {
								store.PlanEndsAt = &planEndsAt.Time
							}
							store.CollectionID = "stores"
							store.CollectionName = "stores"
							stores = append(stores, store)
						}
					}
					expandData["stores"] = stores
				}
			case "addresses":
				var addresses []models.Address
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, line1, line2, city, state, postal_code, country, user FROM addresses WHERE user = ?", 
					user.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var address models.Address
						
						err := rows.Scan(&address.ID, &address.Created, &address.Updated, &address.CollectionID, 
							&address.CollectionName, &address.Line1, &address.Line2, &address.City, 
							&address.State, &address.PostalCode, &address.Country, &address.User)
						if err == nil {
							address.CollectionID = "addresses"
							address.CollectionName = "addresses"
							addresses = append(addresses, address)
						}
					}
					expandData["addresses"] = addresses
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		users[i] = user
	}
	
	return users
}