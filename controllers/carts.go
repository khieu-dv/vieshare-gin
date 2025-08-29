package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Carts handlers for PocketBase compatibility

func (p *PocketBaseController) listCarts(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var carts []models.Cart
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM carts"
	countQuery := "SELECT COUNT(*) FROM carts"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count carts"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch carts"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var cart models.Cart
		var user sql.NullString
		
		err := rows.Scan(&cart.ID, &cart.Created, &cart.Updated, &cart.CollectionID, 
			&cart.CollectionName, &user, &cart.SessionID)
		if err != nil {
			continue
		}
		
		if user.Valid {
			cart.User = user.String
		}
		cart.CollectionID = "carts"
		cart.CollectionName = "carts"
		carts = append(carts, cart)
	}
	
	// Handle expand relations
	if expand != "" {
		carts = p.expandCartRelations(carts, expand)
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
		Items:      carts,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getCart(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var cart models.Cart
	var user sql.NullString
	
	query := "SELECT * FROM carts WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&cart.ID, &cart.Created, &cart.Updated, 
		&cart.CollectionID, &cart.CollectionName, &user, &cart.SessionID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
		}
		return
	}
	
	if user.Valid {
		cart.User = user.String
	}
	cart.CollectionID = "carts"
	cart.CollectionName = "carts"
	
	// Handle expand relations
	if expand != "" {
		carts := []models.Cart{cart}
		carts = p.expandCartRelations(carts, expand)
		if len(carts) > 0 {
			cart = carts[0]
		}
	}
	
	c.JSON(http.StatusOK, cart)
}

func (p *PocketBaseController) createCart(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var cart models.Cart
	setBaseRecord(&cart.BaseRecord, "carts")
	
	// Map data to cart struct
	if user, ok := data["user"].(string); ok && user != "" {
		cart.User = user
	}
	if sessionID, ok := data["session_id"].(string); ok {
		cart.SessionID = sessionID
	}
	
	// Insert into database
	query := `INSERT INTO carts (id, created, updated, collection_id, collection_name, user, session_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	var userValue interface{}
	if cart.User != "" {
		userValue = cart.User
	}
		
	_, err := dbMap.Db.Exec(query, cart.ID, cart.Created, cart.Updated, cart.CollectionID, 
		cart.CollectionName, userValue, cart.SessionID)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
		return
	}
	
	c.JSON(http.StatusOK, cart)
}

func (p *PocketBaseController) updateCart(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if cart exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM carts WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if user, ok := data["user"].(string); ok {
		if user == "" {
			setParts = append(setParts, "user = NULL")
		} else {
			setParts = append(setParts, "user = ?")
			args = append(args, user)
		}
	}
	if sessionID, ok := data["session_id"].(string); ok {
		setParts = append(setParts, "session_id = ?")
		args = append(args, sessionID)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE carts SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
		return
	}
	
	// Return updated cart
	p.getCart(c, id, "")
}

func (p *PocketBaseController) deleteCart(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM carts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandCartRelations(carts []models.Cart, expand string) []models.Cart {
	if expand == "" {
		return carts
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, cart := range carts {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "user":
				if cart.User != "" {
					var user models.User
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, email, email_visibility, username, name, avatar, verified FROM users WHERE id = ?", 
						cart.User).Scan(&user.ID, &user.Created, &user.Updated, &user.CollectionID, 
						&user.CollectionName, &user.Email, &user.EmailVisibility, &user.Username, 
						&user.Name, &user.Avatar, &user.Verified)
					if err == nil {
						user.CollectionID = "users"
						user.CollectionName = "users"
						expandData["user"] = user
					}
				}
			case "cart_items":
				var cartItems []models.CartItem
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, cart, product, quantity, subcategory FROM cart_items WHERE cart = ?", 
					cart.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var cartItem models.CartItem
						
						err := rows.Scan(&cartItem.ID, &cartItem.Created, &cartItem.Updated, &cartItem.CollectionID, 
							&cartItem.CollectionName, &cartItem.Cart, &cartItem.Product, &cartItem.Quantity, &cartItem.Subcategory)
						if err == nil {
							cartItem.CollectionID = "cart_items"
							cartItem.CollectionName = "cart_items"
							cartItems = append(cartItems, cartItem)
						}
					}
					expandData["cart_items"] = cartItems
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		carts[i] = cart
	}
	
	return carts
}