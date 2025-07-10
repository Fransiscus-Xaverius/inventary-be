-- Seeder for banners table

-- Insert sample banners
INSERT INTO banners (title, description, cta_text, cta_link, image_url, order_index, is_active, created_at, updated_at)
VALUES
('Summer Collection', 'Up to 50% Off - Discover our new summer collection with amazing discounts!', 'Shop Now', 'https://example.com/summer-collection', 'https://picsum.photos/id/237/1200/400', 1, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('New Arrivals', 'Fresh Styles - Explore the latest trends and elevate your wardrobe.', 'View Collection', 'https://example.com/new-arrivals', 'https://picsum.photos/id/238/1200/400', 2, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Flash Sale', 'Limited Time Offer - Don''t miss out on our exclusive flash sale!', 'Get Deals', 'https://example.com/flash-sale', 'https://picsum.photos/id/239/1200/400', 3, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Winter Wonderland', 'Cozy Comfort - Stay warm and stylish with our winter collection.', 'Discover More', 'https://example.com/winter-collection', 'https://picsum.photos/id/240/1200/400', 4, FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    cta_text = EXCLUDED.cta_text,
    cta_link = EXCLUDED.cta_link,
    image_url = EXCLUDED.image_url,
    order_index = EXCLUDED.order_index,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;
