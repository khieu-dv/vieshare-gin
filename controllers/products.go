package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Products handlers for PocketBase compatibility

func (p *PocketBaseController) listProducts(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var products []models.Product
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM products"
	countQuery := "SELECT COUNT(*) FROM products"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Add default filter for active products
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var product models.Product
		var imagesJSON string
		
		err := rows.Scan(&product.ID, &product.Created, &product.Updated, &product.CollectionID, &product.CollectionName,
			&product.Name, &product.Description, &imagesJSON, &product.Category, &product.Subcategory,
			&product.Price, &product.Inventory, &product.Rating, &product.Store, &product.Active)
		if err != nil {
			continue
		}
		
		// Parse images JSON
		if err := product.Images.Scan(imagesJSON); err == nil {
			product.CollectionID = "products"
			product.CollectionName = "products"
			products = append(products, product)
		}
	}
	
	// Handle expand relations
	if expand != "" {
		products = p.expandProductRelations(products, expand)
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
		Items:      products,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getProduct(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var product models.Product
	var imagesJSON string
	
	query := "SELECT * FROM products WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&product.ID, &product.Created, &product.Updated, 
		&product.CollectionID, &product.CollectionName, &product.Name, &product.Description, 
		&imagesJSON, &product.Category, &product.Subcategory, &product.Price, 
		&product.Inventory, &product.Rating, &product.Store, &product.Active)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}
	
	// Parse images JSON
	product.Images.Scan(imagesJSON)
	product.CollectionID = "products"
	product.CollectionName = "products"
	
	// Handle expand relations
	if expand != "" {
		products := []models.Product{product}
		products = p.expandProductRelations(products, expand)
		if len(products) > 0 {
			product = products[0]
		}
	}
	
	c.JSON(http.StatusOK, product)
}

func (p *PocketBaseController) createProduct(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var product models.Product
	setBaseRecord(&product.BaseRecord, "products")
	
	// Map data to product struct
	if name, ok := data["name"].(string); ok {
		product.Name = name
	}
	if description, ok := data["description"].(string); ok {
		product.Description = description
	}
	if category, ok := data["category"].(string); ok {
		product.Category = category
	}
	if subcategory, ok := data["subcategory"].(string); ok {
		product.Subcategory = subcategory
	}
	if price, ok := data["price"].(string); ok {
		product.Price = price
	}
	if inventory, ok := data["inventory"].(float64); ok {
		product.Inventory = int(inventory)
	}
	if rating, ok := data["rating"].(float64); ok {
		product.Rating = rating
	}
	if store, ok := data["store"].(string); ok {
		product.Store = store
	}
	if active, ok := data["active"].(bool); ok {
		product.Active = active
	} else {
		product.Active = true
	}
	
	// Handle images array
	if images, ok := data["images"].([]interface{}); ok {
		var imageStrings []string
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				imageStrings = append(imageStrings, imgStr)
			}
		}
		product.Images = imageStrings
	}
	
	// Insert into database
	imagesJSON, _ := product.Images.Value()
	query := `INSERT INTO products (id, created, updated, collection_id, collection_name, name, description, 
		images, category, subcategory, price, inventory, rating, store, active) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err := dbMap.Db.Exec(query, product.ID, product.Created, product.Updated, product.CollectionID, 
		product.CollectionName, product.Name, product.Description, imagesJSON, product.Category, 
		product.Subcategory, product.Price, product.Inventory, product.Rating, product.Store, product.Active)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}
	
	c.JSON(http.StatusOK, product)
}

func (p *PocketBaseController) updateProduct(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if product exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	
	// Build update query dynamically
	var setParts []string
	var args []interface{}
	
	if name, ok := data["name"].(string); ok {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if description, ok := data["description"].(string); ok {
		setParts = append(setParts, "description = ?")
		args = append(args, description)
	}
	if category, ok := data["category"].(string); ok {
		setParts = append(setParts, "category = ?")
		args = append(args, category)
	}
	if subcategory, ok := data["subcategory"].(string); ok {
		setParts = append(setParts, "subcategory = ?")
		args = append(args, subcategory)
	}
	if price, ok := data["price"].(string); ok {
		setParts = append(setParts, "price = ?")
		args = append(args, price)
	}
	if inventory, ok := data["inventory"].(float64); ok {
		setParts = append(setParts, "inventory = ?")
		args = append(args, int(inventory))
	}
	if rating, ok := data["rating"].(float64); ok {
		setParts = append(setParts, "rating = ?")
		args = append(args, rating)
	}
	if store, ok := data["store"].(string); ok {
		setParts = append(setParts, "store = ?")
		args = append(args, store)
	}
	if active, ok := data["active"].(bool); ok {
		setParts = append(setParts, "active = ?")
		args = append(args, active)
	}
	
	// Handle images array
	if images, ok := data["images"].([]interface{}); ok {
		var imageStrings models.StringSlice
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				imageStrings = append(imageStrings, imgStr)
			}
		}
		imagesJSON, _ := imageStrings.Value()
		setParts = append(setParts, "images = ?")
		args = append(args, imagesJSON)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = ?")
	args = append(args, "datetime('now')")
	args = append(args, id)
	
	query := "UPDATE products SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	
	// Return updated product
	p.getProduct(c, id, "")
}

func (p *PocketBaseController) deleteProduct(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandProductRelations(products []models.Product, expand string) []models.Product {
	if expand == "" {
		return products
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, product := range products {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "category":
				if product.Category != "" {
					var category models.Category
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, image FROM categories WHERE id = ?", 
						product.Category).Scan(&category.ID, &category.Created, &category.Updated, &category.CollectionID, 
						&category.CollectionName, &category.Name, &category.Slug, &category.Description, &category.Image)
					if err == nil {
						category.CollectionID = "categories"
						category.CollectionName = "categories"
						expandData["category"] = category
					}
				}
			case "subcategory":
				if product.Subcategory != "" {
					var subcategory models.Subcategory
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, category FROM subcategories WHERE id = ?", 
						product.Subcategory).Scan(&subcategory.ID, &subcategory.Created, &subcategory.Updated, &subcategory.CollectionID, 
						&subcategory.CollectionName, &subcategory.Name, &subcategory.Slug, &subcategory.Description, &subcategory.Category)
					if err == nil {
						subcategory.CollectionID = "subcategories"
						subcategory.CollectionName = "subcategories"
						expandData["subcategory"] = subcategory
					}
				}
			case "store":
				if product.Store != "" {
					var store models.Store
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, user, plan, plan_ends_at, cancel_plan_at_end, product_limit, tag_limit, variant_limit, active FROM stores WHERE id = ?", 
						product.Store).Scan(&store.ID, &store.Created, &store.Updated, &store.CollectionID, 
						&store.CollectionName, &store.Name, &store.Slug, &store.Description, &store.User,
						&store.Plan, &store.PlanEndsAt, &store.CancelPlanAtEnd, &store.ProductLimit,
						&store.TagLimit, &store.VariantLimit, &store.Active)
					if err == nil {
						store.CollectionID = "stores"
						store.CollectionName = "stores"
						expandData["store"] = store
					}
				}
			}
		}
		
		// Convert product to map and add expand data
		productMap := map[string]interface{}{
			"id":             product.ID,
			"created":        product.Created,
			"updated":        product.Updated,
			"collectionId":   product.CollectionID,
			"collectionName": product.CollectionName,
			"name":           product.Name,
			"description":    product.Description,
			"images":         product.Images,
			"category":       product.Category,
			"subcategory":    product.Subcategory,
			"price":          product.Price,
			"inventory":      product.Inventory,
			"rating":         product.Rating,
			"store":          product.Store,
			"active":         product.Active,
		}
		
		if len(expandData) > 0 {
			productMap["expand"] = expandData
		}
		
		products[i] = product // Keep the original structure for now
	}
	
	return products
}