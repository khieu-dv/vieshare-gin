package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// BaseRecord represents the common fields for all PocketBase records
type BaseRecord struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	CollectionID string    `json:"collectionId"`
	CollectionName string  `json:"collectionName"`
}

// StringSlice represents a JSON array of strings
type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("cannot scan into StringSlice")
	}
}

func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// JSONField represents a JSON field
type JSONField map[string]interface{}

func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = JSONField{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.New("cannot scan into JSONField")
	}
}

func (j JSONField) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return json.Marshal(j)
}

// User represents the PocketBase users collection
type User struct {
	BaseRecord
	Email          string `json:"email" gorm:"uniqueIndex;not null"`
	EmailVisibility bool  `json:"emailVisibility" gorm:"default:false"`
	Username       string `json:"username" gorm:"uniqueIndex;not null"`
	Name           string `json:"name"`
	Avatar         string `json:"avatar"`
	Verified       bool   `json:"verified" gorm:"default:false"`
}

func (User) TableName() string {
	return "users"
}

// Category represents the categories collection
type Category struct {
	BaseRecord
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Slug        string `json:"slug" gorm:"uniqueIndex;not null"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func (Category) TableName() string {
	return "categories"
}

// Subcategory represents the subcategories collection
type Subcategory struct {
	BaseRecord
	Name        string `json:"name" gorm:"not null"`
	Slug        string `json:"slug" gorm:"not null"`
	Description string `json:"description"`
	Category    string `json:"category" gorm:"not null"` // relation ID
}

func (Subcategory) TableName() string {
	return "subcategories"
}

// Store represents the stores collection
type Store struct {
	BaseRecord
	Name             string `json:"name" gorm:"not null;size:50"`
	Slug             string `json:"slug" gorm:"uniqueIndex;not null"`
	Description      string `json:"description" gorm:"type:text"`
	User             string `json:"user" gorm:"not null"` // relation ID
	Plan             string `json:"plan" gorm:"default:free"`
	PlanEndsAt       *time.Time `json:"plan_ends_at"`
	CancelPlanAtEnd  bool   `json:"cancel_plan_at_end" gorm:"default:false"`
	ProductLimit     int    `json:"product_limit" gorm:"default:10"`
	TagLimit         int    `json:"tag_limit" gorm:"default:5"`
	VariantLimit     int    `json:"variant_limit" gorm:"default:5"`
	Active           bool   `json:"active" gorm:"default:true"`
}

func (Store) TableName() string {
	return "stores"
}

// Product represents the products collection
type Product struct {
	BaseRecord
	Name        string      `json:"name" gorm:"not null"`
	Description string      `json:"description" gorm:"type:text"`
	Images      StringSlice `json:"images" gorm:"type:text"`
	Category    string      `json:"category" gorm:"not null"` // relation ID
	Subcategory string      `json:"subcategory"`              // relation ID
	Price       string      `json:"price" gorm:"not null"`
	Inventory   int         `json:"inventory" gorm:"default:0"`
	Rating      float64     `json:"rating" gorm:"default:0"`
	Store       string      `json:"store" gorm:"not null"` // relation ID
	Active      bool        `json:"active" gorm:"default:true"`
}

func (Product) TableName() string {
	return "products"
}

// Cart represents the carts collection
type Cart struct {
	BaseRecord
	User      string `json:"user"`       // relation ID (optional for guest carts)
	SessionID string `json:"session_id"` // for guest carts
}

func (Cart) TableName() string {
	return "carts"
}

// CartItem represents the cart_items collection
type CartItem struct {
	BaseRecord
	Cart        string `json:"cart" gorm:"not null"`        // relation ID
	Product     string `json:"product" gorm:"not null"`     // relation ID
	Quantity    int    `json:"quantity" gorm:"not null;min:1"`
	Subcategory string `json:"subcategory"`                 // relation ID
}

func (CartItem) TableName() string {
	return "cart_items"
}

// Address represents the addresses collection
type Address struct {
	BaseRecord
	Line1      string `json:"line1" gorm:"not null"`
	Line2      string `json:"line2"`
	City       string `json:"city" gorm:"not null"`
	State      string `json:"state" gorm:"not null"`
	PostalCode string `json:"postal_code" gorm:"not null"`
	Country    string `json:"country" gorm:"not null"`
	User       string `json:"user" gorm:"not null"` // relation ID
}

func (Address) TableName() string {
	return "addresses"
}

// Order represents the orders collection
type Order struct {
	BaseRecord
	User     string    `json:"user"`           // relation ID (optional for guest orders)
	Store    string    `json:"store" gorm:"not null"` // relation ID
	Items    JSONField `json:"items" gorm:"type:text;not null"`
	Quantity int       `json:"quantity"`
	Amount   string    `json:"amount" gorm:"not null"`
	Status   string    `json:"status" gorm:"default:pending"`
	Name     string    `json:"name" gorm:"not null"`
	Email    string    `json:"email" gorm:"not null"`
	Address  string    `json:"address" gorm:"not null"` // relation ID
	Notes    string    `json:"notes"`
}

func (Order) TableName() string {
	return "orders"
}

// Customer represents the customers collection
type Customer struct {
	BaseRecord
	Name        string `json:"name"`
	Email       string `json:"email" gorm:"not null"`
	Store       string `json:"store" gorm:"not null"` // relation ID
	TotalOrders int    `json:"total_orders" gorm:"default:0"`
	TotalSpent  string `json:"total_spent" gorm:"default:0"`
}

func (Customer) TableName() string {
	return "customers"
}

// Notification represents the notifications collection
type Notification struct {
	BaseRecord
	Email         string `json:"email" gorm:"uniqueIndex;not null"`
	Token         string `json:"token" gorm:"uniqueIndex;not null"`
	User          string `json:"user"` // relation ID
	Communication bool   `json:"communication" gorm:"default:false"`
	Newsletter    bool   `json:"newsletter" gorm:"default:false"`
	Marketing     bool   `json:"marketing" gorm:"default:false"`
}

func (Notification) TableName() string {
	return "notifications"
}

// PocketBase API Response structures
type PBListResponse struct {
	Page         int         `json:"page"`
	PerPage      int         `json:"perPage"`
	TotalItems   int         `json:"totalItems"`
	TotalPages   int         `json:"totalPages"`
	Items        interface{} `json:"items"`
}

type PBAuthResponse struct {
	Token  string      `json:"token"`
	Record interface{} `json:"record"`
}

// Expand structures for relations
type ExpandData map[string]interface{}