-- 1. Products
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    short_description TEXT,
    description TEXT,

    category_id BIGINT NOT NULL,
    brand_id BIGINT NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_listed BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_products_brand
        FOREIGN KEY (brand_id)
        REFERENCES brands(id)
        ON DELETE RESTRICT
);

CREATE INDEX products_category_id_idx
    ON products(category_id);

CREATE INDEX products_brand_id_idx
    ON products(brand_id);

CREATE INDEX products_active_listed_idx
    ON products(is_active, is_listed);

CREATE INDEX products_deleted_at_idx
    ON products(deleted_at);


-- 2. Product variants
CREATE TABLE product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,

    size VARCHAR(50),
    color VARCHAR(100),
    sku VARCHAR(100) NOT NULL UNIQUE,

    mrp NUMERIC(12,2) NOT NULL,
    selling_price NUMERIC(12,2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_product_variants_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE
);

CREATE INDEX product_variants_product_id_idx
    ON product_variants(product_id);

CREATE INDEX product_variants_active_idx
    ON product_variants(is_active);


-- 3. Product images
CREATE TABLE product_images (
    id BIGSERIAL PRIMARY KEY,

    product_id BIGINT NOT NULL,
    variant_id BIGINT,

    image_url VARCHAR(500) NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_product_images_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_product_images_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE SET NULL
);

CREATE INDEX product_images_product_id_idx
    ON product_images(product_id);

CREATE INDEX product_images_variant_id_idx
    ON product_images(variant_id);