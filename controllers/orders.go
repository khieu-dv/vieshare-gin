package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Orders handlers for PocketBase compatibility

func (p *PocketBaseController) listOrders(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var orders []models.Order
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM orders"
	countQuery := "SELECT COUNT(*) FROM orders"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count orders"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var order models.Order
		var itemsJSON string
		var user sql.NullString
		var quantity sql.NullInt64
		
		err := rows.Scan(&order.ID, &order.Created, &order.Updated, &order.CollectionID, 
			&order.CollectionName, &user, &order.Store, &itemsJSON, &quantity, 
			&order.Amount, &order.Status, &order.Name, &order.Email, &order.Address, &order.Notes)
		if err != nil {
			continue
		}
		
		if user.Valid {
			order.User = user.String
		}
		if quantity.Valid {
			order.Quantity = int(quantity.Int64)
		}
		
		// Parse items JSON
		if err := order.Items.Scan(itemsJSON); err == nil {
			order.CollectionID = "orders"
			order.CollectionName = "orders"
			orders = append(orders, order)
		}
	}
	
	// Handle expand relations
	if expand != "" {
		orders = p.expandOrderRelations(orders, expand)
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
		Items:      orders,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getOrder(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var order models.Order
	var itemsJSON string
	var user sql.NullString
	var quantity sql.NullInt64
	
	query := "SELECT * FROM orders WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&order.ID, &order.Created, &order.Updated, 
		&order.CollectionID, &order.CollectionName, &user, &order.Store, &itemsJSON, 
		&quantity, &order.Amount, &order.Status, &order.Name, &order.Email, &order.Address, &order.Notes)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}
	
	if user.Valid {
		order.User = user.String
	}
	if quantity.Valid {
		order.Quantity = int(quantity.Int64)
	}
	
	// Parse items JSON
	order.Items.Scan(itemsJSON)
	order.CollectionID = "orders"
	order.CollectionName = "orders"
	
	// Handle expand relations
	if expand != "" {
		orders := []models.Order{order}
		orders = p.expandOrderRelations(orders, expand)
		if len(orders) > 0 {
			order = orders[0]
		}
	}
	
	c.JSON(http.StatusOK, order)
}

func (p *PocketBaseController) createOrder(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var order models.Order
	setBaseRecord(&order.BaseRecord, "orders")
	
	// Map data to order struct
	if user, ok := data["user"].(string); ok && user != "" {
		order.User = user
	}
	if store, ok := data["store"].(string); ok {
		order.Store = store
	}
	if quantity, ok := data["quantity"].(float64); ok {
		order.Quantity = int(quantity)
	}
	if amount, ok := data["amount"].(string); ok {
		order.Amount = amount
	}
	if status, ok := data["status"].(string); ok {
		order.Status = status
	} else {
		order.Status = "pending"
	}
	if name, ok := data["name"].(string); ok {
		order.Name = name
	}
	if email, ok := data["email"].(string); ok {
		order.Email = email
	}
	if address, ok := data["address"].(string); ok {
		order.Address = address
	}
	if notes, ok := data["notes"].(string); ok {
		order.Notes = notes
	}
	
	// Handle items object/map
	if items, ok := data["items"].(map[string]interface{}); ok {
		order.Items = items
	} else {
		order.Items = make(models.JSONField)
	}
	
	// Insert into database
	itemsJSON, _ := order.Items.Value()
	query := `INSERT INTO orders (id, created, updated, collection_id, collection_name, user, store, 
		items, quantity, amount, status, name, email, address, notes) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	var userValue interface{}
	if order.User != "" {
		userValue = order.User
	}
	var quantityValue interface{}
	if order.Quantity > 0 {
		quantityValue = order.Quantity
	}
		
	_, err := dbMap.Db.Exec(query, order.ID, order.Created, order.Updated, order.CollectionID, 
		order.CollectionName, userValue, order.Store, itemsJSON, quantityValue, order.Amount, 
		order.Status, order.Name, order.Email, order.Address, order.Notes)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}
	
	c.JSON(http.StatusOK, order)
}

func (p *PocketBaseController) updateOrder(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if order exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
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
	if store, ok := data["store"].(string); ok {
		setParts = append(setParts, "store = ?")
		args = append(args, store)
	}
	if quantity, ok := data["quantity"].(float64); ok {
		if quantity == 0 {
			setParts = append(setParts, "quantity = NULL")
		} else {
			setParts = append(setParts, "quantity = ?")
			args = append(args, int(quantity))
		}
	}
	if amount, ok := data["amount"].(string); ok {
		setParts = append(setParts, "amount = ?")
		args = append(args, amount)
	}
	if status, ok := data["status"].(string); ok {
		setParts = append(setParts, "status = ?")
		args = append(args, status)
	}
	if name, ok := data["name"].(string); ok {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if email, ok := data["email"].(string); ok {
		setParts = append(setParts, "email = ?")
		args = append(args, email)
	}
	if address, ok := data["address"].(string); ok {
		setParts = append(setParts, "address = ?")
		args = append(args, address)
	}
	if notes, ok := data["notes"].(string); ok {
		setParts = append(setParts, "notes = ?")
		args = append(args, notes)
	}
	
	// Handle items object/map
	if items, ok := data["items"].(map[string]interface{}); ok {
		var itemsJSON models.JSONField = items
		itemsValue, _ := itemsJSON.Value()
		setParts = append(setParts, "items = ?")
		args = append(args, itemsValue)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE orders SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}
	
	// Return updated order
	p.getOrder(c, id, "")
}

func (p *PocketBaseController) deleteOrder(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM orders WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandOrderRelations(orders []models.Order, expand string) []models.Order {
	if expand == "" {
		return orders
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, order := range orders {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "user":
				if order.User != "" {
					var user models.User
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, email, email_visibility, username, name, avatar, verified FROM users WHERE id = ?", 
						order.User).Scan(&user.ID, &user.Created, &user.Updated, &user.CollectionID, 
						&user.CollectionName, &user.Email, &user.EmailVisibility, &user.Username, 
						&user.Name, &user.Avatar, &user.Verified)
					if err == nil {
						user.CollectionID = "users"
						user.CollectionName = "users"
						expandData["user"] = user
					}
				}
			case "store":
				if order.Store != "" {
					var store models.Store
					var planEndsAt sql.NullTime
					
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, user, plan, plan_ends_at, cancel_plan_at_end, product_limit, tag_limit, variant_limit, active FROM stores WHERE id = ?", 
						order.Store).Scan(&store.ID, &store.Created, &store.Updated, &store.CollectionID, 
						&store.CollectionName, &store.Name, &store.Slug, &store.Description, &store.User,
						&store.Plan, &planEndsAt, &store.CancelPlanAtEnd, &store.ProductLimit,
						&store.TagLimit, &store.VariantLimit, &store.Active)
					if err == nil {
						if planEndsAt.Valid {
							store.PlanEndsAt = &planEndsAt.Time
						}
						store.CollectionID = "stores"
						store.CollectionName = "stores"
						expandData["store"] = store
					}
				}
			case "address":
				if order.Address != "" {
					var address models.Address
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, line1, line2, city, state, postal_code, country, user FROM addresses WHERE id = ?", 
						order.Address).Scan(&address.ID, &address.Created, &address.Updated, &address.CollectionID, 
						&address.CollectionName, &address.Line1, &address.Line2, &address.City, 
						&address.State, &address.PostalCode, &address.Country, &address.User)
					if err == nil {
						address.CollectionID = "addresses"
						address.CollectionName = "addresses"
						expandData["address"] = address
					}
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		orders[i] = order
	}
	
	return orders
}