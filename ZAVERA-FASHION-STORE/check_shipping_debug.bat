@echo off
echo ============================================
echo ZAVERA SHIPPING COST DEBUG CHECKER
echo ============================================
echo.
echo This script helps debug why shipping costs are 2x more expensive
echo.
pause

echo.
echo [STEP 1] Checking if dimension columns exist...
echo.
psql -U postgres -d zavera_db -c "\d products" | findstr "length width height"

echo.
echo [STEP 2] Checking product dimensions in database...
echo.
psql -U postgres -d zavera_db -f database\check_product_dimensions.sql

echo.
echo ============================================
echo ANALYSIS:
echo ============================================
echo.
echo Check the output above:
echo.
echo 1. Do you see "length", "width", "height" columns?
echo    - If NO: Run migrate_product_dimensions.bat
echo    - If YES: Continue to step 2
echo.
echo 2. Check Denim Jacket dimensions:
echo    - Weight should be: 700 (grams per item)
echo    - Length should be: 30 (cm)
echo    - Width should be: 25 (cm)
echo    - Height should be: 10 (cm)
echo.
echo 3. If dimensions are 0 or wrong:
echo    - Go to Admin Dashboard
echo    - Edit the product
echo    - Set correct dimensions
echo    - Save
echo.
echo 4. Test checkout and check backend terminal logs:
echo    - Look for: "Item 1: Denim Jacket - Weight: 700g, Dimensions: 30x25x10 cm"
echo    - If shows "0x0x0", dimensions not loaded from database
echo.
echo 5. IMPORTANT: Biteship uses VOLUMETRIC WEIGHT
echo    - Formula: (L x W x H) / 6000
echo    - For 30x25x10: (30x25x10)/6000 = 1.25kg per item
echo    - Total for 2 items: 2.5kg (volumetric) vs 1.4kg (actual)
echo    - Biteship uses HIGHER value = 2.5kg
echo    - This explains higher shipping cost!
echo.
echo ============================================
echo.
echo Read SHIPPING_COST_DEBUG_GUIDE.md for detailed instructions
echo.
pause
