package controllers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/VieShare/vieshare-gin/db"
	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
)

// Subcategories handlers for PocketBase compatibility

func (p *PocketBaseController) listSubcategories(c *gin.Context, page, perPage, offset int, sort, filter, expand string) {
	dbMap := db.GetDB()
	
	var subcategories []models.Subcategory
	var totalItems int64
	
	// Build query
	baseQuery := "SELECT * FROM subcategories"
	countQuery := "SELECT COUNT(*) FROM subcategories"
	
	// Add filter
	whereClause, args := buildFilterClause(filter)
	if whereClause != "" {
		baseQuery += " " + whereClause
		countQuery += " " + whereClause
	}
	
	// Get total count
	err := dbMap.Db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count subcategories"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subcategories"})
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var subcategory models.Subcategory
		
		err := rows.Scan(&subcategory.ID, &subcategory.Created, &subcategory.Updated, &subcategory.CollectionID, 
			&subcategory.CollectionName, &subcategory.Name, &subcategory.Slug, &subcategory.Description, &subcategory.Category)
		if err != nil {
			continue
		}
		
		subcategory.CollectionID = "subcategories"
		subcategory.CollectionName = "subcategories"
		subcategories = append(subcategories, subcategory)
	}
	
	// Handle expand relations
	if expand != "" {
		subcategories = p.expandSubcategoryRelations(subcategories, expand)
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
		Items:      subcategories,
	}
	
	c.JSON(http.StatusOK, response)
}

func (p *PocketBaseController) getSubcategory(c *gin.Context, id, expand string) {
	dbMap := db.GetDB()
	
	var subcategory models.Subcategory
	
	query := "SELECT * FROM subcategories WHERE id = ?"
	err := dbMap.Db.QueryRow(query, id).Scan(&subcategory.ID, &subcategory.Created, &subcategory.Updated, 
		&subcategory.CollectionID, &subcategory.CollectionName, &subcategory.Name, &subcategory.Slug, 
		&subcategory.Description, &subcategory.Category)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subcategory not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subcategory"})
		}
		return
	}
	
	subcategory.CollectionID = "subcategories"
	subcategory.CollectionName = "subcategories"
	
	// Handle expand relations
	if expand != "" {
		subcategories := []models.Subcategory{subcategory}
		subcategories = p.expandSubcategoryRelations(subcategories, expand)
		if len(subcategories) > 0 {
			subcategory = subcategories[0]
		}
	}
	
	c.JSON(http.StatusOK, subcategory)
}

func (p *PocketBaseController) createSubcategory(c *gin.Context, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	var subcategory models.Subcategory
	setBaseRecord(&subcategory.BaseRecord, "subcategories")
	
	// Map data to subcategory struct
	if name, ok := data["name"].(string); ok {
		subcategory.Name = name
	}
	if slug, ok := data["slug"].(string); ok {
		subcategory.Slug = slug
	}
	if description, ok := data["description"].(string); ok {
		subcategory.Description = description
	}
	if category, ok := data["category"].(string); ok {
		subcategory.Category = category
	}
	
	// Insert into database
	query := `INSERT INTO subcategories (id, created, updated, collection_id, collection_name, name, slug, 
		description, category) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
	_, err := dbMap.Db.Exec(query, subcategory.ID, subcategory.Created, subcategory.Updated, subcategory.CollectionID, 
		subcategory.CollectionName, subcategory.Name, subcategory.Slug, subcategory.Description, subcategory.Category)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subcategory"})
		return
	}
	
	c.JSON(http.StatusOK, subcategory)
}

func (p *PocketBaseController) updateSubcategory(c *gin.Context, id string, data map[string]interface{}) {
	dbMap := db.GetDB()
	
	// First check if subcategory exists
	var exists bool
	err := dbMap.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM subcategories WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subcategory not found"})
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
	if category, ok := data["category"].(string); ok {
		setParts = append(setParts, "category = ?")
		args = append(args, category)
	}
	
	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	
	// Always update the updated timestamp
	setParts = append(setParts, "updated = datetime('now')")
	args = append(args, id)
	
	query := "UPDATE subcategories SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	
	_, err = dbMap.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subcategory"})
		return
	}
	
	// Return updated subcategory
	p.getSubcategory(c, id, "")
}

func (p *PocketBaseController) deleteSubcategory(c *gin.Context, id string) {
	dbMap := db.GetDB()
	
	result, err := dbMap.Db.Exec("DELETE FROM subcategories WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subcategory"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subcategory not found"})
		return
	}
	
	c.JSON(http.StatusNoContent, nil)
}

func (p *PocketBaseController) expandSubcategoryRelations(subcategories []models.Subcategory, expand string) []models.Subcategory {
	if expand == "" {
		return subcategories
	}
	
	dbMap := db.GetDB()
	expandFields := strings.Split(expand, ",")
	
	for i, subcategory := range subcategories {
		expandData := make(map[string]interface{})
		
		for _, field := range expandFields {
			field = strings.TrimSpace(field)
			
			switch field {
			case "category":
				if subcategory.Category != "" {
					var category models.Category
					err := dbMap.Db.QueryRow("SELECT id, created, updated, collection_id, collection_name, name, slug, description, image FROM categories WHERE id = ?", 
						subcategory.Category).Scan(&category.ID, &category.Created, &category.Updated, &category.CollectionID, 
						&category.CollectionName, &category.Name, &category.Slug, &category.Description, &category.Image)
					if err == nil {
						category.CollectionID = "categories"
						category.CollectionName = "categories"
						expandData["category"] = category
					}
				}
			case "products":
				var products []models.Product
				rows, err := dbMap.Db.Query("SELECT id, created, updated, collection_id, collection_name, name, description, images, category, subcategory, price, inventory, rating, store, active FROM products WHERE subcategory = ? AND active = true", 
					subcategory.ID)
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
		subcategories[i] = subcategory
	}
	
	return subcategories
}