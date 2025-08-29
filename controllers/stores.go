package controllers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Stores handlers for PocketBase compatibility

func (p *PocketBaseController) listStores(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var stores []models.Store
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM stores"
	countQuery := "SELECT COUNT(*) FROM stores"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Add default filter for active stores
	if !strings.Contains(filter, "active") {
		if whereClause != "" {
			baseQuery += " AND active = true"
			countQuery += " AND active = true"
		} else {
			baseQuery += " WHERE active = true"
			countQuery += " WHERE active = true"
		}
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count stores"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stores"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var store models.Store
		var planEndsAt sql.NullTime
		
		err := rows.Scan(&store.ID, &store.Created, &store.Updated, &store.CollectionID, &store.CollectionName,
			&store.Name, &store.Slug, &store.Description, &store.User, &store.Plan, &planEndsAt,
			&store.CancelPlanAtEnd, &store.ProductLimit, &store.TagLimit, &store.VariantLimit, &store.Active)
		if err != nil {
			continue
		}
		
		if planEndsAt.Valid {
			store.PlanEndsAt = &planEndsAt.Time
		}
		store.CollectionID = "stores"
		store.CollectionName = "stores"
		stores = append(stores, store)
	}
	
	// Handle expand relations
	if expand != "" {
		stores = p.expandStoreRelations(stores, expand)
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
		Items:      stores,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getStore(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var store models.Store
	var planEndsAt sql.NullTime
	
	query := "SELECT * FROM stores WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&store.ID, &store.Created, &store.Updated, 
		&store.CollectionID, &store.CollectionName, &store.Name, &store.Slug, &store.Description, 
		&store.User, &store.Plan, &planEndsAt, &store.CancelPlanAtEnd, &store.ProductLimit,
		&store.TagLimit, &store.VariantLimit, &store.Active)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch store"})
		}
		return
	}
	
	if planEndsAt.Valid {
		store.PlanEndsAt = &planEndsAt.Time
	}
	store.CollectionID = "stores"
	store.CollectionName = "stores"
	
	// Handle expand relations
	if expand != "" {
		stores := []models.Store{store}
		stores = p.expandStoreRelations(stores, expand)
		if len(stores) > 0 {
			store = stores[0]
		}
	}
	
	c.JSON(http.StatusOK, store)
}

func (p *PocketBaseController) createStore(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var store models.Store
	setBaseRecord(&store.BaseRecord, "stores")
	
	// Map data to store struct
	if name, ok := data["name"].(string); ok {
		store.Name = name
	}
	if slug, ok := data["slug"].(string); ok {
		store.Slug = slug
	}
	if description, ok := data["description"].(string); ok {
		store.Description = description
	}
	if user, ok := data["user"].(string); ok {
		store.User = user
	}
	if plan, ok := data["plan"].(string); ok {
		store.Plan = plan
	} else {
		store.Plan = "free"
	}
	if planEndsAt, ok := data["plan_ends_at"].(string); ok && planEndsAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, planEndsAt); err == nil {
			store.PlanEndsAt = &parsedTime
		}
	}
	if cancelPlanAtEnd, ok := data["cancel_plan_at_end"].(bool); ok {
		store.CancelPlanAtEnd = cancelPlanAtEnd
	}
	if productLimit, ok := data["product_limit"].(float64); ok {
		store.ProductLimit = int(productLimit)
	} else {
		store.ProductLimit = 10
	}
	if tagLimit, ok := data["tag_limit"].(float64); ok {
		store.TagLimit = int(tagLimit)
	} else {
		store.TagLimit = 5
	}
	if variantLimit, ok := data["variant_limit"].(float64); ok {
		store.VariantLimit = int(variantLimit)
	} else {
		store.VariantLimit = 5
	}
	if active, ok := data["active"].(bool); ok {
		store.Active = active
	} else {
		store.Active = true
	}
	
	// Insert into database
	query := `INSERT INTO stores (id, created, updated, collection_id, collection_name, name, slug, description, 
		user, plan, plan_ends_at, cancel_plan_at_end, product_limit, tag_limit, variant_limit, active) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	var planEndsAtValue interface{}
	if store.PlanEndsAt != nil {
		planEndsAtValue = store.PlanEndsAt
	}
		
	_, err := dbMap.Db.Exec(query, store.ID, store.Created, store.Updated, store.CollectionID, 
		store.CollectionName, store.Name, store.Slug, store.Description, store.User, store.Plan,
		planEndsAtValue, store.CancelPlanAtEnd, store.ProductLimit, store.TagLimit, 
		store.VariantLimit, store.Active)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create store"})
		return
	}
	
	c.JSON(http.StatusOK, store)
}

func (p *PocketBaseController) updateStore(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if store exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM stores WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if name, ok := data["name"].(string); ok {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if slug, ok := data["slug"].(string); ok {
		setParts = append(setParts, "slug = ?")
		args = append(args, slug)
	}
	if description, ok := data["description"].(string); ok {
		setParts = append(setParts, "description = ?")
		args = append(args, description)
	}
	if user, ok := data["user"].(string); ok {
		setParts = append(setParts, "user = ?")
		args = append(args, user)
	}
	if plan, ok := data["plan"].(string); ok {
		setParts = append(setParts, "plan = ?")
		args = append(args, plan)
	}
	if planEndsAt, ok := data["plan_ends_at"].(string); ok {
		if planEndsAt == "" {
			setParts = append(setParts, "plan_ends_at = NULL")
		} else {
			if parsedTime, err := time.Parse(time.RFC3339, planEndsAt); err == nil {
				setParts = append(setParts, "plan_ends_at = ?")
				args = append(args, parsedTime)
			}
		}
	}
	if cancelPlanAtEnd, ok := data["cancel_plan_at_end"].(bool); ok {
		setParts = append(setParts, "cancel_plan_at_end = ?")
		args = append(args, cancelPlanAtEnd)
	}
	if productLimit, ok := data["product_limit"].(float64); ok {
		setParts = append(setParts, "product_limit = ?")
		args = append(args, int(productLimit))
	}
	if tagLimit, ok := data["tag_limit"].(float64); ok {
		setParts = append(setParts, "tag_limit = ?")
		args = append(args, int(tagLimit))
	}
	if variantLimit, ok := data["variant_limit"].(float64); ok {
		setParts = append(setParts, "variant_limit = ?")
		args = append(args, int(variantLimit))
	}
	if active, ok := data["active"].(bool); ok {
		setParts = append(setParts, "active = ?")
		args = append(args, active)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE stores SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store"})
		return
	}
	
	// Return updated store
	p.getStore(c, id, "")
}

func (p *PocketBaseController) deleteStore(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM stores WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete store"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandStoreRelations(stores []models.Store, expand string) []models.Store {
	if expand == "" {
		return stores
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, store := range stores {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "user":
				if store.User != "" {
					var user models.User
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, email, email_visibility, username, name, avatar, verified FROM users WHERE id = ?", 
						store.User).Scan(&user.ID, &user.Created, &user.Updated, &user.CollectionID, 
						&user.CollectionName, &user.Email, &user.EmailVisibility, &user.Username, 
						&user.Name, &user.Avatar, &user.Verified)
					if err == nil {
						user.CollectionID = "users"
						user.CollectionName = "users"
						expandData["user"] = user
					}
				}
			case "products":
				var products []models.Product
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, name, description, images, category, subcategory, price, inventory, rating, store, active FROM products WHERE store = ? AND active = true", 
					store.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var product models.Product
						var imagesJSON string
						
						err := rows.Scan(&product.ID, &product.Created, &product.Updated, &product.CollectionID, 
							&product.CollectionName, &product.Name, &product.Description, &imagesJSON, &product.Category, 
							&product.Subcategory, &product.Price, &product.Inventory, &product.Rating, &product.Store, &product.Active)
						if err == nil {
							product.Images.Scan(imagesJSON)
							product.CollectionID = "products"
							product.CollectionName = "products"
							products = append(products, product)
						}
					}
					expandData["products"] = products
				}
			case "orders":
				var orders []models.Order
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, user, store, items, quantity, amount, status, name, email, address, notes FROM orders WHERE store = ?", 
					store.ID)
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
			case "customers":
				var customers []models.Customer
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, name, email, store, total_orders, total_spent FROM customers WHERE store = ?", 
					store.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var customer models.Customer
						
						err := rows.Scan(&customer.ID, &customer.Created, &customer.Updated, &customer.CollectionID, 
							&customer.CollectionName, &customer.Name, &customer.Email, &customer.Store, 
							&customer.TotalOrders, &customer.TotalSpent)
						if err == nil {
							customer.CollectionID = "customers"
							customer.CollectionName = "customers"
							customers = append(customers, customer)
						}
					}
					expandData["customers"] = customers
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		stores[i] = store
	}
	
	return stores
}