![VieShare Gin API](https://upload.wikimedia.org/wikipedia/commons/2/23/Golang.png)

# VieShare Gin Private API

Welcome to **VieShare Gin Private API** - A PocketBase-compatible backend API built with [Gin Framework](https://github.com/gin-gonic/gin/) for the VieShare e-commerce platform.

This API provides a complete backend solution with **SQLite** database, **PocketBase-compatible** endpoints, **JWT** authentication, and **Redis** caching support.

## Features

### 🚀 **PocketBase-Compatible API**
- Complete REST API endpoints matching PocketBase format: `/api/collections/{collection}/records`
- Support for all CRUD operations (GET, POST, PATCH, DELETE)
- Advanced filtering, sorting, and pagination
- Relation expansion with `expand` parameter
- Health check endpoint at `/api/health`

### 🏪 **E-commerce Collections**
- **Users**: User authentication and profiles
- **Categories & Subcategories**: Product categorization
- **Products**: Product catalog with images, pricing, inventory
- **Stores**: Multi-vendor store management
- **Carts & Cart Items**: Shopping cart functionality
- **Orders**: Order processing and management
- **Addresses**: Shipping address management
- **Customers**: Customer tracking and analytics
- **Notifications**: User notification preferences

### 🔧 **Technical Stack**
- [Gin Framework](https://github.com/gin-gonic/gin/): High-performance HTTP web framework
- [go-gorp](https://github.com/go-gorp/gorp): Go Relational Persistence (ORM)
- [SQLite3](https://github.com/mattn/go-sqlite3): Embedded database with auto-initialization
- [jwt-go](https://github.com/golang-jwt/jwt): JSON Web Tokens for authentication
- [go-redis](https://github.com/go-redis/redis): Redis caching support
- [Swagger](https://github.com/swaggo/gin-swagger): API documentation
- Built-in **CORS Middleware** for cross-origin requests
- Built-in **RequestID Middleware** for request tracking
- **Environment-based configuration**
- **SSL/TLS support**

## Installation

### Prerequisites
- Go 1.19 or higher
- Git

### Download FrontEnd, please checkout [Gin Framework](https://github.com/khieu-dv/vieshare.git)


### 1. Clone the Repository


```bash
git clone https://github.com/khieu-dv/vieshare-gin.git
cd vieshare-gin
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Database Setup
The application uses SQLite with auto-initialization. The database schema is automatically created from `db/pocketbase_schema.sql` on first run.

**Sample data included:**
- 4 categories (Vieboards, Clothing, Shoes, Accessories)
- 9 subcategories (Decks, Wheels, T-shirts, etc.)
- 3 sample products
- 1 admin user and store

### 4. Environment Configuration

Create and configure your `.env` file:
```bash
cp .env_rename_me .env
```

**Required environment variables:**
```env
# Server Configuration
PORT=9000
ENV=LOCAL

# Database Configuration
DB_PATH=./data/app.db

```

## Running the Application

### Development Mode
```bash
go run *.go
```

### Build and Run
```bash
go build -o vieshare-gin .
./vieshare-gin
```

### With SSL (Optional)
Generate SSL certificates:
```bash
mkdir cert/
sh generate-certificate.sh
```
Set `SSL=TRUE` in your `.env` file.

## API Documentation

### PocketBase-Compatible Endpoints

The API provides PocketBase-compatible endpoints that can be used with existing PocketBase SDKs or direct HTTP calls:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/health` | Health check |
| `GET` | `/api/collections/{collection}/records` | List records |
| `GET` | `/api/collections/{collection}/records/{id}` | Get single record |
| `POST` | `/api/collections/{collection}/records` | Create record |
| `PATCH` | `/api/collections/{collection}/records/{id}` | Update record |
| `DELETE` | `/api/collections/{collection}/records/{id}` | Delete record |

### Available Collections

- `users` - User accounts and authentication
- `categories` - Product categories
- `subcategories` - Product subcategories  
- `stores` - Merchant stores
- `products` - Product catalog
- `carts` - Shopping carts
- `cart_items` - Cart items
- `addresses` - Shipping addresses
- `orders` - Order management
- `customers` - Customer data
- `notifications` - User notifications

### Query Parameters

- `page` - Page number (default: 1)
- `perPage` - Records per page (default: 30, max: 100)
- `sort` - Sort fields (e.g., `created,-updated`)
- `filter` - Filter query (e.g., `active=true`)
- `expand` - Expand relations (e.g., `category,store`)

### Example Requests

```bash
# Get all products with pagination
curl "http://localhost:9000/api/collections/products/records?page=1&perPage=10"

# Get product with expanded category and store
curl "http://localhost:9000/api/collections/products/records/prod_deck_001?expand=category,store"

# Filter active products by category
curl "http://localhost:9000/api/collections/products/records?filter=active=true&&category=cat_vieboards"
```

### Response Format

All responses follow the PocketBase format:

```json
{
  "page": 1,
  "perPage": 30,
  "totalItems": 100,
  "totalPages": 4,
  "items": [
    {
      "id": "record_id",
      "created": "2023-01-01T00:00:00Z",
      "updated": "2023-01-01T00:00:00Z",
      "collectionId": "collection_name",
      "collectionName": "collection_name",
      // ... record fields
    }
  ]
}
```

## Swagger Documentation

Generate and view API documentation:
```bash
make generate_docs
make run
open http://localhost:9000/swagger/index.html
```

## Legacy Authentication API

The application also maintains legacy JWT authentication endpoints for backward compatibility:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/user/login` | User login |
| `POST` | `/v1/user/register` | User registration |
| `GET` | `/v1/user/logout` | User logout |
| `POST` | `/v1/token/refresh` | Refresh JWT token |

## Project Structure

```
vieshare-gin/
├── controllers/          # API controllers
│   ├── pocketbase.go    # Main PocketBase controller
│   ├── products.go      # Products collection
│   ├── users.go         # Users collection
│   └── ...              # Other collections
├── db/                  # Database layer
│   ├── db.go           # Database connection
│   └── pocketbase_schema.sql  # Database schema
├── models/              # Data models
│   ├── pocketbase.go   # PocketBase-compatible models
│   └── user.go         # Legacy user model
├── forms/              # Form validators
├── public/             # Static files
├── .env               # Environment configuration
└── main.go            # Application entry point
```


