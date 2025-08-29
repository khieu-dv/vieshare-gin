package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Categories handlers for PocketBase compatibility

func (p *PocketBaseController) listCategories(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var categories []models.Category
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM categories"
	countQuery := "SELECT COUNT(*) FROM categories"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count categories"})
		return
	}
	
	// Add sorting
	sortClause := buildSortClause(sort)
	if sortClause == "" {
		sortClause = "ORDER BY name ASC"
	}
	baseQuery += " " + sortClause
	
	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, perPage, offset)
	
	// Execute query
	rows, err := dbMap.Db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var category models.Category
		
		err := rows.Scan(&category.ID, &category.Created, &category.Updated, &category.CollectionID, 
			&category.CollectionName, &category.Name, &category.Slug, &category.Description, &category.Image)
		if err != nil {
			continue
		}
		
		category.CollectionID = "categories"
		category.CollectionName = "categories"
		categories = append(categories, category)
	}
	
	// Handle expand relations
	if expand != "" {
		categories = p.expandCategoryRelations(categories, expand)
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
		Items:      categories,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getCategory(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var category models.Category
	
	query := "SELECT * FROM categories WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&category.ID, &category.Created, &category.Updated, 
		&category.CollectionID, &category.CollectionName, &category.Name, &category.Slug, 
		&category.Description, &category.Image)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category"})
		}
		return
	}
	
	category.CollectionID = "categories"
	category.CollectionName = "categories"
	
	// Handle expand relations
	if expand != "" {
		categories := []models.Category{category}
		categories = p.expandCategoryRelations(categories, expand)
		if len(categories) > 0 {
			category = categories[0]
		}
	}
	
	c.JSON(http.StatusOK, category)
}

func (p *PocketBaseController) createCategory(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var category models.Category
	setBaseRecord(&category.BaseRecord, "categories")
	
	// Map data to category struct
	if name, ok := data["name"].(string); ok {
		category.Name = name
	}
	if slug, ok := data["slug"].(string); ok {
		category.Slug = slug
	}
	if description, ok := data["description"].(string); ok {
		category.Description = description
	}
	if image, ok := data["image"].(string); ok {
		category.Image = image
	}
	
	// Insert into database
	query := `INSERT INTO categories (id, created, updated, collection_id, collection_name, name, slug, 
		description, image) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err := dbMap.Db.Exec(query, category.ID, category.Created, category.Updated, category.CollectionID, 
		category.CollectionName, category.Name, category.Slug, category.Description, category.Image)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}
	
	c.JSON(http.StatusOK, category)
}

func (p *PocketBaseController) updateCategory(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if category exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
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
	if image, ok := data["image"].(string); ok {
		setParts = append(setParts, "image = ?")
		args = append(args, image)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE categories SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}
	
	// Return updated category
	p.getCategory(c, id, "")
}

func (p *PocketBaseController) deleteCategory(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandCategoryRelations(categories []models.Category, expand string) []models.Category {
	if expand == "" {
		return categories
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, category := range categories {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "subcategories":
				var subcategories []models.Subcategory
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, name, slug, description, category FROM subcategories WHERE category = ?", 
					category.ID)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var subcategory models.Subcategory
						
						err := rows.Scan(&subcategory.ID, &subcategory.Created, &subcategory.Updated, &subcategory.CollectionID, 
							&subcategory.CollectionName, &subcategory.Name, &subcategory.Slug, &subcategory.Description, &subcategory.Category)
						if err == nil {
							subcategory.CollectionID = "subcategories"
							subcategory.CollectionName = "subcategories"
							subcategories = append(subcategories, subcategory)
						}
					}
					expandData["subcategories"] = subcategories
				}
			case "products":
				var products []models.Product
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, name, description, images, category, subcategory, price, inventory, rating, store, active FROM products WHERE category = ? AND active = true", 
					category.ID)
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
			}
		}
		
		// For now, we'll just store the expand data somehow
		// In a real implementation, you'd want to modify the struct or use a different approach
		categories[i] = category
	}
	
	return categories
}