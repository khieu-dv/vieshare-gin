package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Cart Items handlers for PocketBase compatibility

func (p *PocketBaseController) listCartItems(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var cartItems []models.CartItem
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM cart_items"
	countQuery := "SELECT COUNT(*) FROM cart_items"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count cart items"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart items"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var cartItem models.CartItem
		
		err := rows.Scan(&cartItem.ID, &cartItem.Created, &cartItem.Updated, &cartItem.CollectionID, 
			&cartItem.CollectionName, &cartItem.Cart, &cartItem.Product, &cartItem.Quantity, &cartItem.Subcategory)
		if err != nil {
			continue
		}
		
		cartItem.CollectionID = "cart_items"
		cartItem.CollectionName = "cart_items"
		cartItems = append(cartItems, cartItem)
	}
	
	// Handle expand relations
	if expand != "" {
		cartItems = p.expandCartItemRelations(cartItems, expand)
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
		Items:      cartItems,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getCartItem(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var cartItem models.CartItem
	
	query := "SELECT * FROM cart_items WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&cartItem.ID, &cartItem.Created, &cartItem.Updated, 
		&cartItem.CollectionID, &cartItem.CollectionName, &cartItem.Cart, &cartItem.Product, 
		&cartItem.Quantity, &cartItem.Subcategory)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart item"})
		}
		return
	}
	
	cartItem.CollectionID = "cart_items"
	cartItem.CollectionName = "cart_items"
	
	// Handle expand relations
	if expand != "" {
		cartItems := []models.CartItem{cartItem}
		cartItems = p.expandCartItemRelations(cartItems, expand)
		if len(cartItems) > 0 {
			cartItem = cartItems[0]
		}
	}
	
	c.JSON(http.StatusOK, cartItem)
}

func (p *PocketBaseController) createCartItem(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var cartItem models.CartItem
	setBaseRecord(&cartItem.BaseRecord, "cart_items")
	
	// Map data to cart item struct
	if cart, ok := data["cart"].(string); ok {
		cartItem.Cart = cart
	}
	if product, ok := data["product"].(string); ok {
		cartItem.Product = product
	}
	if quantity, ok := data["quantity"].(float64); ok {
		cartItem.Quantity = int(quantity)
	}
	if subcategory, ok := data["subcategory"].(string); ok {
		cartItem.Subcategory = subcategory
	}
	
	// Check if item already exists
	var existingID string
	checkQuery := "SELECT id FROM cart_items WHERE cart = ? AND product = ?"
	err := dbMap.Db.QueryRow(checkQuery, cartItem.Cart, cartItem.Product).Scan(&existingID)
	
	if err == nil {
		// Item exists, update quantity
		updateQuery := "UPDATE cart_items SET quantity = quantity + ?, updated = datetime('now') WHERE id = ?"
		_, err = dbMap.Db.Exec(updateQuery, cartItem.Quantity, existingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
		
		// Return updated item
		p.getCartItem(c, existingID, "")
		return
	}
	
	// Insert new cart item
	query := `INSERT INTO cart_items (id, created, updated, collection_id, collection_name, cart, product, quantity, subcategory) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err = dbMap.Db.Exec(query, cartItem.ID, cartItem.Created, cartItem.Updated, cartItem.CollectionID, 
		cartItem.CollectionName, cartItem.Cart, cartItem.Product, cartItem.Quantity, cartItem.Subcategory)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart item"})
		return
	}
	
	c.JSON(http.StatusOK, cartItem)
}

func (p *PocketBaseController) updateCartItem(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if cart item exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM cart_items WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if cart, ok := data["cart"].(string); ok {
		setParts = append(setParts, "cart = ?")
		args = append(args, cart)
	}
	if product, ok := data["product"].(string); ok {
		setParts = append(setParts, "product = ?")
		args = append(args, product)
	}
	if quantity, ok := data["quantity"].(float64); ok {
		setParts = append(setParts, "quantity = ?")
		args = append(args, int(quantity))
	}
	if subcategory, ok := data["subcategory"].(string); ok {
		setParts = append(setParts, "subcategory = ?")
		args = append(args, subcategory)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE cart_items SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
		return
	}
	
	// Return updated cart item
	p.getCartItem(c, id, "")
}

func (p *PocketBaseController) deleteCartItem(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM cart_items WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart item"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandCartItemRelations(cartItems []models.CartItem, expand string) []models.CartItem {
	if expand == "" {
		return cartItems
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, cartItem := range cartItems {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "product":
				if cartItem.Product != "" {
					var product models.Product
					var imagesJSON string
					
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, description, images, category, subcategory, price, inventory, rating, store, active FROM products WHERE id = ?", 
						cartItem.Product).Scan(&product.ID, &product.Created, &product.Updated, &product.CollectionID, 
						&product.CollectionName, &product.Name, &product.Description, &imagesJSON, &product.Category, 
						&product.Subcategory, &product.Price, &product.Inventory, &product.Rating, &product.Store, &product.Active)
					
					if err == nil {
						product.Images.Scan(imagesJSON)
						product.CollectionID = "products"
						product.CollectionName = "products"
						expandData["product"] = product
					}
				}
			case "cart":
				if cartItem.Cart != "" {
					var cart models.Cart
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, user, session_id FROM carts WHERE id = ?", 
						cartItem.Cart).Scan(&cart.ID, &cart.Created, &cart.Updated, &cart.CollectionID, 
						&cart.CollectionName, &cart.User, &cart.SessionID)
					if err == nil {
						cart.CollectionID = "carts"
						cart.CollectionName = "carts"
						expandData["cart"] = cart
					}
				}
			case "subcategory":
				if cartItem.Subcategory != "" {
					var subcategory models.Subcategory
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, category FROM subcategories WHERE id = ?", 
						cartItem.Subcategory).Scan(&subcategory.ID, &subcategory.Created, &subcategory.Updated, &subcategory.CollectionID, 
						&subcategory.CollectionName, &subcategory.Name, &subcategory.Slug, &subcategory.Description, &subcategory.Category)
					if err == nil {
						subcategory.CollectionID = "subcategories"
						subcategory.CollectionName = "subcategories"
						expandData["subcategory"] = subcategory
					}
				}
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		cartItems[i] = cartItem
	}
	
	return cartItems
}