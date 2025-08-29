package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Customers handlers for PocketBase compatibility

func (p *PocketBaseController) listCustomers(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var customers []models.Customer
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM customers"
	countQuery := "SELECT COUNT(*) FROM customers"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count customers"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var customer models.Customer
		
		err := rows.Scan(&customer.ID, &customer.Created, &customer.Updated, &customer.CollectionID, 
			&customer.CollectionName, &customer.Name, &customer.Email, &customer.Store, 
			&customer.TotalOrders, &customer.TotalSpent)
		if err != nil {
			continue
		}
		
		customer.CollectionID = "customers"
		customer.CollectionName = "customers"
		customers = append(customers, customer)
	}
	
	// Handle expand relations
	if expand != "" {
		customers = p.expandCustomerRelations(customers, expand)
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
		Items:      customers,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getCustomer(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var customer models.Customer
	
	query := "SELECT * FROM customers WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&customer.ID, &customer.Created, &customer.Updated, 
		&customer.CollectionID, &customer.CollectionName, &customer.Name, &customer.Email, 
		&customer.Store, &customer.TotalOrders, &customer.TotalSpent)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer"})
		}
		return
	}
	
	customer.CollectionID = "customers"
	customer.CollectionName = "customers"
	
	// Handle expand relations
	if expand != "" {
		customers := []models.Customer{customer}
		customers = p.expandCustomerRelations(customers, expand)
		if len(customers) > 0 {
			customer = customers[0]
		}
	}
	
	c.JSON(http.StatusOK, customer)
}

func (p *PocketBaseController) createCustomer(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var customer models.Customer
	setBaseRecord(&customer.BaseRecord, "customers")
	
	// Map data to customer struct
	if name, ok := data["name"].(string); ok {
		customer.Name = name
	}
	if email, ok := data["email"].(string); ok {
		customer.Email = email
	}
	if store, ok := data["store"].(string); ok {
		customer.Store = store
	}
	if totalOrders, ok := data["total_orders"].(float64); ok {
		customer.TotalOrders = int(totalOrders)
	}
	if totalSpent, ok := data["total_spent"].(string); ok {
		customer.TotalSpent = totalSpent
	} else {
		customer.TotalSpent = "0"
	}
	
	// Insert into database
	query := `INSERT INTO customers (id, created, updated, collection_id, collection_name, name, email, 
		store, total_orders, total_spent) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err := dbMap.Db.Exec(query, customer.ID, customer.Created, customer.Updated, customer.CollectionID, 
		customer.CollectionName, customer.Name, customer.Email, customer.Store, customer.TotalOrders, customer.TotalSpent)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}
	
	c.JSON(http.StatusOK, customer)
}

func (p *PocketBaseController) updateCustomer(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if customer exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM customers WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if name, ok := data["name"].(string); ok {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if email, ok := data["email"].(string); ok {
		setParts = append(setParts, "email = ?")
		args = append(args, email)
	}
	if store, ok := data["store"].(string); ok {
		setParts = append(setParts, "store = ?")
		args = append(args, store)
	}
	if totalOrders, ok := data["total_orders"].(float64); ok {
		setParts = append(setParts, "total_orders = ?")
		args = append(args, int(totalOrders))
	}
	if totalSpent, ok := data["total_spent"].(string); ok {
		setParts = append(setParts, "total_spent = ?")
		args = append(args, totalSpent)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE customers SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}
	
	// Return updated customer
	p.getCustomer(c, id, "")
}

func (p *PocketBaseController) deleteCustomer(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM customers WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandCustomerRelations(customers []models.Customer, expand string) []models.Customer {
	if expand == "" {
		return customers
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, customer := range customers {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "store":
				if customer.Store != "" {
					var store models.Store
					var planEndsAt sql.NullTime
					
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, user, plan, plan_ends_at, cancel_plan_at_end, product_limit, tag_limit, variant_limit, active FROM stores WHERE id = ?", 
						customer.Store).Scan(&store.ID, &store.Created, &store.Updated, &store.CollectionID, 
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
			case "orders":
				var orders []models.Order
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, user, store, items, quantity, amount, status, name, email, address, notes FROM orders WHERE email = ?", 
					customer.Email)
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
		customers[i] = customer
	}
	
	return customers
}