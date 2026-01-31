@echo off
echo ========================================
echo TEST CATEGORY MAPPING
echo ========================================
echo.

echo Step 1: Check database for test product
echo ----------------------------------------
psql -U postgres -d zavera_db -c "SELECT id, name, category, subcategory FROM products WHERE name LIKE '%%Test%%' ORDER BY id DESC LIMIT 5;"

echo.
echo Step 2: Check all distinct subcategories
echo ----------------------------------------
psql -U postgres -d zavera_db -c "SELECT DISTINCT category, subcategory FROM products ORDER BY category, subcategory;"

echo.
echo Step 3: Verify Pria category mapping
echo ----------------------------------------
psql -U postgres -d zavera_db -c "SELECT id, name, subcategory, CASE WHEN subcategory = 'Tops' THEN 'Atasan' WHEN subcategory = 'Shirts' THEN 'Kemeja' WHEN subcategory = 'Bottoms' THEN 'Celana' WHEN subcategory = 'Outerwear' THEN 'Jaket' WHEN subcategory = 'Suits' THEN 'Jas' WHEN subcategory = 'Footwear' THEN 'Sepatu' ELSE subcategory END as client_display FROM products WHERE category = 'pria' ORDER BY id DESC LIMIT 10;"

echo.
echo ========================================
echo Test Instructions:
echo ========================================
echo 1. Go to: http://localhost:3000/admin/products/add
echo 2. Create product:
echo    - Category: Pria
echo    - Subcategory: Atasan
echo    - Name: Test Atasan Mapping
echo 3. Check database (run this script again)
echo 4. Go to: http://localhost:3000/pria
echo 5. Filter by "Atasan"
echo 6. Product should appear!
echo.
pause
