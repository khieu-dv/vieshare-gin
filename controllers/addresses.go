package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Addresses handlers for PocketBase compatibility

func (p *PocketBaseController) listAddresses(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var addresses []models.Address
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM addresses"
	countQuery := "SELECT COUNT(*) FROM addresses"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count addresses"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch addresses"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var address models.Address
		
		err := rows.Scan(&address.ID, &address.Created, &address.Updated, &address.CollectionID, 
			&address.CollectionName, &address.Line1, &address.Line2, &address.City, 
			&address.State, &address.PostalCode, &address.Country, &address.User)
		if err != nil {
			continue
		}
		
		address.CollectionID = "addresses"
		address.CollectionName = "addresses"
		addresses = append(addresses, address)
	}
	
	// Handle expand relations
	if expand != "" {
		addresses = p.expandAddressRelations(addresses, expand)
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
		Items:      addresses,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getAddress(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var address models.Address
	
	query := "SELECT * FROM addresses WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&address.ID, &address.Created, &address.Updated, 
		&address.CollectionID, &address.CollectionName, &address.Line1, &address.Line2, 
		&address.City, &address.State, &address.PostalCode, &address.Country, &address.User)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch address"})
		}
		return
	}
	
	address.CollectionID = "addresses"
	address.CollectionName = "addresses"
	
	// Handle expand relations
	if expand != "" {
		addresses := []models.Address{address}
		addresses = p.expandAddressRelations(addresses, expand)
		if len(addresses) > 0 {
			address = addresses[0]
		}
	}
	
	c.JSON(http.StatusOK, address)
}

func (p *PocketBaseController) createAddress(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var address models.Address
	setBaseRecord(&address.BaseRecord, "addresses")
	
	// Map data to address struct
	if line1, ok := data["line1"].(string); ok {
		address.Line1 = line1
	}
	if line2, ok := data["line2"].(string); ok {
		address.Line2 = line2
	}
	if city, ok := data["city"].(string); ok {
		address.City = city
	}
	if state, ok := data["state"].(string); ok {
		address.State = state
	}
	if postalCode, ok := data["postal_code"].(string); ok {
		address.PostalCode = postalCode
	}
	if country, ok := data["country"].(string); ok {
		address.Country = country
	}
	if user, ok := data["user"].(string); ok {
		address.User = user
	}
	
	// Insert into database
	query := `INSERT INTO addresses (id, created, updated, collection_id, collection_name, line1, line2, 
		city, state, postal_code, country, user) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err := dbMap.Db.Exec(query, address.ID, address.Created, address.Updated, address.CollectionID, 
		address.CollectionName, address.Line1, address.Line2, address.City, address.State, 
		address.PostalCode, address.Country, address.User)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create address"})
		return
	}
	
	c.JSON(http.StatusOK, address)
}

func (p *PocketBaseController) updateAddress(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if address exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM addresses WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if line1, ok := data["line1"].(string); ok {
		setParts = append(setParts, "line1 = ?")
		args = append(args, line1)
	}
	if line2, ok := data["line2"].(string); ok {
		setParts = append(setParts, "line2 = ?")
		args = append(args, line2)
	}
	if city, ok := data["city"].(string); ok {
		setParts = append(setParts, "city = ?")
		args = append(args, city)
	}
	if state, ok := data["state"].(string); ok {
		setParts = append(setParts, "state = ?")
		args = append(args, state)
	}
	if postalCode, ok := data["postal_code"].(string); ok {
		setParts = append(setParts, "postal_code = ?")
		args = append(args, postalCode)
	}
	if country, ok := data["country"].(string); ok {
		setParts = append(setParts, "country = ?")
		args = append(args, country)
	}
	if user, ok := data["user"].(string); ok {
		setParts = append(setParts, "user = ?")
		args = append(args, user)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE addresses SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		return
	}
	
	// Return updated address
	p.getAddress(c, id, "")
}

func (p *PocketBaseController) deleteAddress(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM addresses WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandAddressRelations(addresses []models.Address, expand string) []models.Address {
	if expand == "" {
		return addresses
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, address := range addresses {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "user":
				if address.User != "" {
					var user models.User
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, email, email_visibility, username, name, avatar, verified FROM users WHERE id = ?", 
						address.User).Scan(&user.ID, &user.Created, &user.Updated, &user.CollectionID, 
						&user.CollectionName, &user.Email, &user.EmailVisibility, &user.Username, 
						&user.Name, &user.Avatar, &user.Verified)
					if err == nil {
						user.CollectionID = "users"
						user.CollectionName = "users"
						expandData["user"] = user
					}
				}
			case "orders":
				var orders []models.Order
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, user, store, items, quantity, amount, status, name, email, address, notes FROM orders WHERE address = ?", 
					address.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var order models.Order
						var itemsJSON string
						var orderUser sql.NullString
						var quantity sql.NullInt64
						
						err := rows.Scan(&order.ID, &order.Created, &order.Updated, &order.CollectionID, 
							&order.CollectionName, &orderUser, &order.Store, &itemsJSON, &quantity, 
							&order.Amount, &order.Status, &order.Name, &order.Email, &order.Address, &order.Notes)
						if err == nil {
							if orderUser.Valid {
								order.User = orderUser.String
							}
							if quantity.Valid {
								order.Quantity = int(quantity.Int64)
							}
							order.Items.Scan(itemsJSON)
							order.CollectionID = "orders"
							order.CollectionName = "orders"
							orders = append(orders, order)
						}
					}
					expandData["orders"] = orders
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		addresses[i] = address
	}
	
	return addresses
}