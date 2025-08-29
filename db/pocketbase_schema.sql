-- PocketBase-compatible schema for VieShare
-- Drop existing tables
DROP TABLE IF EXISTS article;
DROP TABLE IF EXISTS user;

-- Create new PocketBase-compatible tables

-- Users table (PocketBase built-in auth)
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'users',
    collection_name TEXT DEFAULT 'users',
    email TEXT NOT NULL UNIQUE,
    email_visibility BOOLEAN DEFAULT FALSE,
    username TEXT NOT NULL UNIQUE,
    name TEXT,
    avatar TEXT,
    verified BOOLEAN DEFAULT FALSE
);

-- Categories table
CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'categories',
    collection_name TEXT DEFAULT 'categories',
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    image TEXT
);

-- Subcategories table
CREATE TABLE subcategories (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'subcategories',
    collection_name TEXT DEFAULT 'subcategories',
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL,
    FOREIGN KEY (category) REFERENCES categories(id) ON DELETE CASCADE
);

-- Stores table
CREATE TABLE stores (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'stores',
    collection_name TEXT DEFAULT 'stores',
    name TEXT NOT NULL CHECK(LENGTH(name) <= 50),
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    user TEXT NOT NULL,
    plan TEXT DEFAULT 'free',
    plan_ends_at DATETIME,
    cancel_plan_at_end BOOLEAN DEFAULT FALSE,
    product_limit INTEGER DEFAULT 10,
    tag_limit INTEGER DEFAULT 5,
    variant_limit INTEGER DEFAULT 5,
    active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (user) REFERENCES users(id) ON DELETE CASCADE
);

-- Products table
CREATE TABLE products (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'products',
    collection_name TEXT DEFAULT 'products',
    name TEXT NOT NULL,
    description TEXT,
    images TEXT DEFAULT '[]',
    category TEXT NOT NULL,
    subcategory TEXT,
    price TEXT NOT NULL,
    inventory INTEGER DEFAULT 0,
    rating REAL DEFAULT 0,
    store TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (category) REFERENCES categories(id) ON DELETE CASCADE,
    FOREIGN KEY (subcategory) REFERENCES subcategories(id) ON DELETE SET NULL,
    FOREIGN KEY (store) REFERENCES stores(id) ON DELETE CASCADE
);

-- Carts table
CREATE TABLE carts (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'carts',
    collection_name TEXT DEFAULT 'carts',
    user TEXT,
    session_id TEXT,
    FOREIGN KEY (user) REFERENCES users(id) ON DELETE CASCADE
);

-- Cart Items table
CREATE TABLE cart_items (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'cart_items',
    collection_name TEXT DEFAULT 'cart_items',
    cart TEXT NOT NULL,
    product TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK(quantity >= 1),
    subcategory TEXT,
    FOREIGN KEY (cart) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (product) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (subcategory) REFERENCES subcategories(id) ON DELETE SET NULL,
    UNIQUE(cart, product)
);

-- Addresses table
CREATE TABLE addresses (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'addresses',
    collection_name TEXT DEFAULT 'addresses',
    line1 TEXT NOT NULL,
    line2 TEXT,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    country TEXT NOT NULL,
    user TEXT NOT NULL,
    FOREIGN KEY (user) REFERENCES users(id) ON DELETE CASCADE
);

-- Orders table
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'orders',
    collection_name TEXT DEFAULT 'orders',
    user TEXT,
    store TEXT NOT NULL,
    items TEXT NOT NULL DEFAULT '{}',
    quantity INTEGER,
    amount TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    address TEXT NOT NULL,
    notes TEXT,
    FOREIGN KEY (user) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (store) REFERENCES stores(id) ON DELETE CASCADE,
    FOREIGN KEY (address) REFERENCES addresses(id) ON DELETE RESTRICT
);

-- Customers table
CREATE TABLE customers (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'customers',
    collection_name TEXT DEFAULT 'customers',
    name TEXT,
    email TEXT NOT NULL,
    store TEXT NOT NULL,
    total_orders INTEGER DEFAULT 0,
    total_spent TEXT DEFAULT '0',
    FOREIGN KEY (store) REFERENCES stores(id) ON DELETE CASCADE
);

-- Notifications table
CREATE TABLE notifications (
    id TEXT PRIMARY KEY,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    collection_id TEXT DEFAULT 'notifications',
    collection_name TEXT DEFAULT 'notifications',
    email TEXT NOT NULL UNIQUE,
    token TEXT NOT NULL UNIQUE,
    user TEXT,
    communication BOOLEAN DEFAULT FALSE,
    newsletter BOOLEAN DEFAULT FALSE,
    marketing BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_subcategories_category ON subcategories(category);
CREATE INDEX idx_stores_user ON stores(user);
CREATE INDEX idx_stores_slug ON stores(slug);
CREATE INDEX idx_products_store ON products(store);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_subcategory ON products(subcategory);
CREATE INDEX idx_products_active ON products(active);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_cart_items_cart ON cart_items(cart);
CREATE INDEX idx_cart_items_product ON cart_items(product);
CREATE INDEX idx_orders_user ON orders(user);
CREATE INDEX idx_orders_store ON orders(store);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created);
CREATE INDEX idx_customers_store ON customers(store);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_addresses_user ON addresses(user);
CREATE INDEX idx_carts_user ON carts(user);
CREATE INDEX idx_carts_session_id ON carts(session_id);

-- Create triggers for updating timestamps
CREATE TRIGGER update_users_updated_at 
AFTER UPDATE ON users FOR EACH ROW 
BEGIN 
    UPDATE users SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_categories_updated_at 
AFTER UPDATE ON categories FOR EACH ROW 
BEGIN 
    UPDATE categories SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_subcategories_updated_at 
AFTER UPDATE ON subcategories FOR EACH ROW 
BEGIN 
    UPDATE subcategories SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_stores_updated_at 
AFTER UPDATE ON stores FOR EACH ROW 
BEGIN 
    UPDATE stores SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_products_updated_at 
AFTER UPDATE ON products FOR EACH ROW 
BEGIN 
    UPDATE products SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_carts_updated_at 
AFTER UPDATE ON carts FOR EACH ROW 
BEGIN 
    UPDATE carts SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_cart_items_updated_at 
AFTER UPDATE ON cart_items FOR EACH ROW 
BEGIN 
    UPDATE cart_items SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_addresses_updated_at 
AFTER UPDATE ON addresses FOR EACH ROW 
BEGIN 
    UPDATE addresses SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_orders_updated_at 
AFTER UPDATE ON orders FOR EACH ROW 
BEGIN 
    UPDATE orders SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_customers_updated_at 
AFTER UPDATE ON customers FOR EACH ROW 
BEGIN 
    UPDATE customers SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

CREATE TRIGGER update_notifications_updated_at 
AFTER UPDATE ON notifications FOR EACH ROW 
BEGIN 
    UPDATE notifications SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id; 
END;

-- Insert sample data
INSERT INTO categories (id, name, slug, description) VALUES 
('cat_vieboards', 'Vieboards', 'vieboards', 'The best vieboards for all levels of viers.'),
('cat_clothing', 'Clothing', 'clothing', 'Skateboarding apparel and accessories.'),
('cat_shoes', 'Shoes', 'shoes', 'Skateboarding shoes and footwear.'),
('cat_accessories', 'Accessories', 'accessories', 'Skateboarding accessories and gear.');

INSERT INTO subcategories (id, name, slug, description, category) VALUES 
('subcat_decks', 'Decks', 'decks', 'Skateboard decks', 'cat_vieboards'),
('subcat_wheels', 'Wheels', 'wheels', 'Skateboard wheels', 'cat_vieboards'),
('subcat_bearings', 'Bearings', 'bearings', 'Skateboard bearings', 'cat_vieboards'),
('subcat_trucks', 'Trucks', 'trucks', 'Skateboard trucks', 'cat_vieboards'),
('subcat_tshirts', 'T-shirts', 't-shirts', 'Skateboarding t-shirts', 'cat_clothing'),
('subcat_hoodies', 'Hoodies', 'hoodies', 'Skateboarding hoodies', 'cat_clothing'),
('subcat_sneakers', 'Sneakers', 'sneakers', 'Skateboarding sneakers', 'cat_shoes'),
('subcat_bags', 'Bags', 'bags', 'Skateboarding bags', 'cat_accessories'),
('subcat_helmets', 'Helmets', 'helmets', 'Skateboarding helmets', 'cat_accessories');

-- Sample user
INSERT INTO users (id, email, username, name, verified) VALUES 
('user_sample_123', 'admin@vieshare.com', 'admin', 'VieShare Admin', TRUE);

-- Sample store
INSERT INTO stores (id, name, slug, description, user) VALUES 
('store_sample_123', 'VieShare Store', 'vieshare-store', 'Official VieShare skateboarding store', 'user_sample_123');

-- Sample products
INSERT INTO products (id, name, description, images, category, subcategory, price, inventory, rating, store, active) VALUES 
('prod_deck_001', 'Street Vieboard Deck', 'High-quality maple deck perfect for street skating', 
 '["deck-1.webp", "deck-2.webp"]', 'cat_vieboards', 'subcat_decks', '59.99', 25, 4.5, 'store_sample_123', TRUE),
('prod_wheels_001', 'Pro Skateboard Wheels', 'Premium urethane wheels for smooth rides', 
 '["wheels-1.webp", "wheels-2.webp"]', 'cat_vieboards', 'subcat_wheels', '29.99', 50, 4.2, 'store_sample_123', TRUE),
('prod_tshirt_001', 'VieShare Logo T-Shirt', 'Comfortable cotton t-shirt with VieShare logo', 
 '["tshirt-1.webp", "tshirt-2.webp"]', 'cat_clothing', 'subcat_tshirts', '19.99', 100, 4.0, 'store_sample_123', TRUE);