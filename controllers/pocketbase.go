package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/VieShare/vieshare-gin/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PocketBaseController handles PocketBase-compatible API endpoints
type PocketBaseController struct{}

// Health godoc
// @Summary Health check
// @Description Returns health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (p *PocketBaseController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// ListRecords godoc
// @Summary List records from collection
// @Description Get paginated list of records from specified collection
// @Tags collections
// @Produce json
// @Param collection path string true "Collection name"
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Records per page" default(30)
// @Param sort query string false "Sort fields"
// @Param filter query string false "Filter query"
// @Param expand query string false "Expand relations"
// @Success 200 {object} models.PBListResponse
// @Router /api/collections/{collection}/records [get]
func (p *PocketBaseController) ListRecords(c *gin.Context) {
	collection := c.Param("collection")
	
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "30"))
	
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}
	
	offset := (page - 1) * perPage
	
	// Parse other parameters
	sort := c.Query("sort")
	filter := c.Query("filter")
	expand := c.Query("expand")
	
	// Route to appropriate handler based on collection
	switch collection {
	case "users":
		p.listUsers(c, page, perPage, offset, sort, filter, expand)
	case "categories":
		p.listCategories(c, page, perPage, offset, sort, filter, expand)
	case "subcategories":
		p.listSubcategories(c, page, perPage, offset, sort, filter, expand)
	case "stores":
		p.listStores(c, page, perPage, offset, sort, filter, expand)
	case "products":
		p.listProducts(c, page, perPage, offset, sort, filter, expand)
	case "carts":
		p.listCarts(c, page, perPage, offset, sort, filter, expand)
	case "cart_items":
		p.listCartItems(c, page, perPage, offset, sort, filter, expand)
	case "addresses":
		p.listAddresses(c, page, perPage, offset, sort, filter, expand)
	case "orders":
		p.listOrders(c, page, perPage, offset, sort, filter, expand)
	case "customers":
		p.listCustomers(c, page, perPage, offset, sort, filter, expand)
	case "notifications":
		p.listNotifications(c, page, perPage, offset, sort, filter, expand)
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Collection '%s' not found", collection),
		})
	}
}

// GetRecord godoc
// @Summary Get single record
// @Description Get a single record by ID from specified collection
// @Tags collections
// @Produce json
// @Param collection path string true "Collection name"
// @Param id path string true "Record ID"
// @Param expand query string false "Expand relations"
// @Success 200 {object} map[string]interface{}
// @Router /api/collections/{collection}/records/{id} [get]
func (p *PocketBaseController) GetRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")
	expand := c.Query("expand")
	
	// Route to appropriate handler based on collection
	switch collection {
	case "users":
		p.getUser(c, id, expand)
	case "categories":
		p.getCategory(c, id, expand)
	case "subcategories":
		p.getSubcategory(c, id, expand)
	case "stores":
		p.getStore(c, id, expand)
	case "products":
		p.getProduct(c, id, expand)
	case "carts":
		p.getCart(c, id, expand)
	case "cart_items":
		p.getCartItem(c, id, expand)
	case "addresses":
		p.getAddress(c, id, expand)
	case "orders":
		p.getOrder(c, id, expand)
	case "customers":
		p.getCustomer(c, id, expand)
	case "notifications":
		p.getNotification(c, id, expand)
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Collection '%s' not found", collection),
		})
	}
}

// CreateRecord godoc
// @Summary Create record
// @Description Create a new record in specified collection
// @Tags collections
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param body body map[string]interface{} true "Record data"
// @Success 200 {object} map[string]interface{}
// @Router /api/collections/{collection}/records [post]
func (p *PocketBaseController) CreateRecord(c *gin.Context) {
	collection := c.Param("collection")
	
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON data",
		})
		return
	}
	
	// Route to appropriate handler based on collection
	switch collection {
	case "users":
		p.createUser(c, data)
	case "categories":
		p.createCategory(c, data)
	case "subcategories":
		p.createSubcategory(c, data)
	case "stores":
		p.createStore(c, data)
	case "products":
		p.createProduct(c, data)
	case "carts":
		p.createCart(c, data)
	case "cart_items":
		p.createCartItem(c, data)
	case "addresses":
		p.createAddress(c, data)
	case "orders":
		p.createOrder(c, data)
	case "customers":
		p.createCustomer(c, data)
	case "notifications":
		p.createNotification(c, data)
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Collection '%s' not found", collection),
		})
	}
}

// UpdateRecord godoc
// @Summary Update record
// @Description Update an existing record in specified collection
// @Tags collections
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param id path string true "Record ID"
// @Param body body map[string]interface{} true "Record data"
// @Success 200 {object} map[string]interface{}
// @Router /api/collections/{collection}/records/{id} [patch]
func (p *PocketBaseController) UpdateRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")
	
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON data",
		})
		return
	}
	
	// Route to appropriate handler based on collection
	switch collection {
	case "users":
		p.updateUser(c, id, data)
	case "categories":
		p.updateCategory(c, id, data)
	case "subcategories":
		p.updateSubcategory(c, id, data)
	case "stores":
		p.updateStore(c, id, data)
	case "products":
		p.updateProduct(c, id, data)
	case "carts":
		p.updateCart(c, id, data)
	case "cart_items":
		p.updateCartItem(c, id, data)
	case "addresses":
		p.updateAddress(c, id, data)
	case "orders":
		p.updateOrder(c, id, data)
	case "customers":
		p.updateCustomer(c, id, data)
	case "notifications":
		p.updateNotification(c, id, data)
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Collection '%s' not found", collection),
		})
	}
}

// DeleteRecord godoc
// @Summary Delete record
// @Description Delete a record from specified collection
// @Tags collections
// @Produce json
// @Param collection path string true "Collection name"
// @Param id path string true "Record ID"
// @Success 204
// @Router /api/collections/{collection}/records/{id} [delete]
func (p *PocketBaseController) DeleteRecord(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")
	
	// Route to appropriate handler based on collection
	switch collection {
	case "users":
		p.deleteUser(c, id)
	case "categories":
		p.deleteCategory(c, id)
	case "subcategories":
		p.deleteSubcategory(c, id)
	case "stores":
		p.deleteStore(c, id)
	case "products":
		p.deleteProduct(c, id)
	case "carts":
		p.deleteCart(c, id)
	case "cart_items":
		p.deleteCartItem(c, id)
	case "addresses":
		p.deleteAddress(c, id)
	case "orders":
		p.deleteOrder(c, id)
	case "customers":
		p.deleteCustomer(c, id)
	case "notifications":
		p.deleteNotification(c, id)
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Collection '%s' not found", collection),
		})
	}
}

// Helper functions

func generateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:15]
}

func setBaseRecord(record *models.BaseRecord, collectionName string) {
	now := time.Now()
	if record.ID == "" {
		record.ID = generateID()
	}
	if record.Created.IsZero() {
		record.Created = now
	}
	record.Updated = now
	record.CollectionID = collectionName
	record.CollectionName = collectionName
}

func buildSortClause(sort string) string {
	if sort == "" {
		return ""
	}
	
	sortFields := strings.Split(sort, ",")
	var clauses []string
	
	for _, field := range sortFields {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "-") {
			clauses = append(clauses, field[1:]+" DESC")
		} else if strings.HasPrefix(field, "+") {
			clauses = append(clauses, field[1:]+" ASC")
		} else {
			clauses = append(clauses, field+" ASC")
		}
	}
	
	return "ORDER BY " + strings.Join(clauses, ", ")
}

func buildFilterClause(filter string) (string, []interface{}) {
	if filter == "" {
		return "", nil
	}
	
	// This is a simplified filter parser
	// In production, you'd need a more robust parser for PocketBase filter syntax
	parts := strings.Split(filter, "&&")
	var conditions []string
	var args []interface{}
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "=") {
			keyValue := strings.SplitN(part, "=", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.Trim(strings.TrimSpace(keyValue[1]), "\"'")
				conditions = append(conditions, key+" = ?")
				args = append(args, value)
			}
		}
	}
	
	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND "), args
	}
	
	return "", nil
}