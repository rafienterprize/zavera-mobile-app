import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:carousel_slider/carousel_slider.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../models/product.dart';
import '../services/api_service.dart';
import '../widgets/product_card.dart';
import '../widgets/category_card.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ApiService _apiService = ApiService();
  List<Product> _products = [];
  bool _isLoading = true;

  final List<Map<String, String>> _banners = [
    {
      'image': 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1600&q=80',
      'badge': 'FEATURED',
      'title': 'Luxury Collection',
      'subtitle': 'Koleksi eksklusif dari brand designer ternama dengan kualitas premium',
      'button': 'EXPLORE LUXURY',
    },
    {
      'image': 'https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=1600&q=80',
      'badge': 'KOLEKSI WANITA',
      'title': 'Elegant Style',
      'subtitle': 'Tampil percaya diri dengan koleksi fashion wanita terkini',
      'button': 'SHOP NOW',
    },
    {
      'image': 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=1600&q=80',
      'badge': 'KOLEKSI PRIA',
      'title': 'Modern Gentleman',
      'subtitle': 'Gaya maskulin untuk pria modern dan berkelas',
      'button': 'SHOP NOW',
    },
  ];

  @override
  void initState() {
    super.initState();
    _loadProducts();
  }

  Future<void> _loadProducts() async {
    setState(() => _isLoading = true);
    final products = await _apiService.getProducts();
    setState(() {
      _products = products;
      _isLoading = false;
    });
  }

  Widget _buildNavItem(String label) {
    return Padding(
      padding: const EdgeInsets.only(right: 24),
      child: GestureDetector(
        onTap: () {
          // Navigate to category
        },
        child: Text(
          label,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w500,
            letterSpacing: 0.5,
            color: Color(0xFF1a1a1a),
          ),
        ),
      ),
    );
  }
}

// Horizontal Category Scroll Widget with Auto-Scroll
class HorizontalCategoryScroll extends StatefulWidget {
  const HorizontalCategoryScroll({super.key});

  @override
  State<HorizontalCategoryScroll> createState() => _HorizontalCategoryScrollState();
}

class _HorizontalCategoryScrollState extends State<HorizontalCategoryScroll> {
  final ScrollController _scrollController = ScrollController();
  
  final List<Map<String, String>> categories = [
    {
      'name': 'Wanita',
      'subtitle': 'Koleksi Elegant',
      'image': 'https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=800&q=80',
    },
    {
      'name': 'Pria',
      'subtitle': 'Gaya Maskulin',
      'image': 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=800&q=80',
    },
    {
      'name': 'Sports',
      'subtitle': 'Activewear Premium',
      'image': 'https://images.unsplash.com/photo-1461896836934-ffe607ba8211?w=800&q=80',
    },
    {
      'name': 'Anak',
      'subtitle': 'Fashion Stylish',
      'image': 'https://images.unsplash.com/photo-1503944583220-79d8926ad5e2?w=800&q=80',
    },
    {
      'name': 'Luxury',
      'subtitle': 'Koleksi Eksklusif',
      'image': 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1600&q=80',
    },
    {
      'name': 'Beauty',
      'subtitle': 'Perawatan Premium',
      'image': 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=800&q=80',
    },
  ];

  @override
  void initState() {
    super.initState();
    _startAutoScroll();
  }

  void _startAutoScroll() {
    Future.delayed(const Duration(seconds: 2), () {
      if (!mounted) return;
      _autoScroll();
    });
  }

  void _autoScroll() async {
    if (!mounted) return;
    
    const scrollDuration = Duration(milliseconds: 30000); // 30 seconds for full scroll
    final maxScroll = _scrollController.position.maxScrollExtent;
    
    await _scrollController.animateTo(
      maxScroll,
      duration: scrollDuration,
      curve: Curves.linear,
    );
    
    if (!mounted) return;
    
    // Reset to start and repeat
    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;
    
    await _scrollController.animateTo(
      0,
      duration: const Duration(milliseconds: 1000),
      curve: Curves.easeInOut,
    );
    
    if (!mounted) return;
    
    // Repeat
    await Future.delayed(const Duration(seconds: 2));
    if (mounted) _autoScroll();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 200,
      child: ListView.builder(
        controller: _scrollController,
        scrollDirection: Axis.horizontal,
        itemCount: categories.length,
        itemBuilder: (context, index) {
          final category = categories[index];
          return Padding(
            padding: EdgeInsets.only(
              right: 16,
              left: index == 0 ? 0 : 0,
            ),
            child: GestureDetector(
              onTap: () {
                // Navigate to category
              },
              child: Container(
                width: 160,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.15),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: Stack(
                    fit: StackFit.expand,
                    children: [
                      CachedNetworkImage(
                        imageUrl: category['image']!,
                        fit: BoxFit.cover,
                      ),
                      Container(
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [
                              Colors.black.withValues(alpha: 0.2),
                              Colors.black.withValues(alpha: 0.6),
                            ],
                          ),
                        ),
                      ),
                      Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              category['name']!,
                              style: GoogleFonts.playfairDisplay(
                                fontSize: 22,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                                letterSpacing: 1,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              category['subtitle']!,
                              style: const TextStyle(
                                fontSize: 12,
                                color: Colors.white70,
                              ),
                            ),
                            const SizedBox(height: 12),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                border: Border.all(color: Colors.white, width: 1.5),
                                borderRadius: BorderRadius.circular(2),
                              ),
                              child: const Text(
                                'SHOP NOW',
                                style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 10,
                                  fontWeight: FontWeight.w600,
                                  letterSpacing: 1.2,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            // Header
            Container(
              color: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Column(
                children: [
                  // Logo centered
                  Text(
                    'ZAVERA',
                    style: GoogleFonts.playfairDisplay(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 4,
                      color: const Color(0xFF1a1a1a),
                    ),
                  ),
                  const SizedBox(height: 12),
                  // Search Bar with icons
                  Row(
                    children: [
                      // Search Bar
                      Expanded(
                        child: Container(
                          height: 42,
                          decoration: BoxDecoration(
                            color: Colors.grey[200],
                            borderRadius: BorderRadius.circular(21),
                          ),
                          child: Row(
                            children: [
                              Padding(
                                padding: const EdgeInsets.only(left: 16, right: 8),
                                child: Icon(
                                  Icons.search,
                                  color: Colors.grey[600],
                                  size: 20,
                                ),
                              ),
                              Expanded(
                                child: TextField(
                                  decoration: InputDecoration(
                                    hintText: 'Cari produk, tren, dan merek...',
                                    hintStyle: TextStyle(
                                      fontSize: 13,
                                      color: Colors.grey[600],
                                    ),
                                    border: InputBorder.none,
                                    contentPadding: const EdgeInsets.only(bottom: 10),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      // Profile Icon
                      IconButton(
                        icon: const Icon(Icons.person_outline, size: 24),
                        onPressed: () {},
                        color: const Color(0xFF1a1a1a),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                      const SizedBox(width: 16),
                      // Cart Icon
                      IconButton(
                        icon: const Icon(Icons.shopping_bag_outlined, size: 24),
                        onPressed: () {},
                        color: const Color(0xFF1a1a1a),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            // Category Navigation
            Container(
              color: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 12),
              child: SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  children: [
                    _buildNavItem('WANITA'),
                    _buildNavItem('PRIA'),
                    _buildNavItem('SPORTS'),
                    _buildNavItem('ANAK'),
                    _buildNavItem('LUXURY'),
                    _buildNavItem('BEAUTY'),
                  ],
                ),
              ),
            ),
            // Content
            Expanded(
              child: RefreshIndicator(
                onRefresh: _loadProducts,
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
              // Hero Carousel
              CarouselSlider(
                options: CarouselOptions(
                  height: 450,
                  viewportFraction: 1.0,
                  autoPlay: true,
                  autoPlayInterval: const Duration(seconds: 5),
                  autoPlayCurve: Curves.easeInOut,
                ),
                items: _banners.map((banner) {
                  return Stack(
                    fit: StackFit.expand,
                    children: [
                      CachedNetworkImage(
                        imageUrl: banner['image']!,
                        fit: BoxFit.cover,
                      ),
                      Container(
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [
                              Colors.black.withValues(alpha: 0.3),
                              Colors.black.withValues(alpha: 0.7),
                            ],
                          ),
                        ),
                      ),
                      Positioned(
                        bottom: 60,
                        left: 24,
                        right: 24,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 6,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.white.withValues(alpha: 0.2),
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Text(
                                banner['badge']!,
                                style: const TextStyle(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w600,
                                  color: Colors.white,
                                  letterSpacing: 1.5,
                                ),
                              ),
                            ),
                            const SizedBox(height: 12),
                            Text(
                              banner['title']!,
                              style: GoogleFonts.playfairDisplay(
                                fontSize: 36,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                                height: 1.2,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              banner['subtitle']!,
                              style: const TextStyle(
                                fontSize: 14,
                                color: Colors.white,
                                height: 1.5,
                              ),
                              maxLines: 2,
                            ),
                            const SizedBox(height: 20),
                            ElevatedButton(
                              onPressed: () {},
                              style: ElevatedButton.styleFrom(
                                backgroundColor: Colors.white,
                                foregroundColor: Colors.black,
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 32,
                                  vertical: 14,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(2),
                                ),
                                elevation: 0,
                              ),
                              child: Text(
                                banner['button']!,
                                style: const TextStyle(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w600,
                                  letterSpacing: 1,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  );
                }).toList(),
              ),

              // Categories Section
              Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Jelajahi Kategori',
                      style: GoogleFonts.playfairDisplay(
                        fontSize: 28,
                        fontWeight: FontWeight.bold,
                        color: const Color(0xFF1a1a1a),
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Temukan koleksi fashion terbaik untuk setiap gaya dan kebutuhan Anda',
                      style: TextStyle(
                        fontSize: 14,
                        color: Colors.grey[600],
                        height: 1.5,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 20),
                    const HorizontalCategoryScroll(),
                  ],
                ),
              ),

              // New Arrivals
              Container(
                color: Colors.grey[50],
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'New Arrivals',
                              style: GoogleFonts.playfairDisplay(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const Text(
                              'Koleksi terbaru',
                              style: TextStyle(color: Colors.grey),
                            ),
                          ],
                        ),
                        TextButton(
                          onPressed: () {},
                          child: const Text('Lihat Semua'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _isLoading
                        ? const Center(child: CircularProgressIndicator())
                        : GridView.builder(
                            shrinkWrap: true,
                            physics: const NeverScrollableScrollPhysics(),
                            gridDelegate:
                                const SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: 2,
                              childAspectRatio: 0.65,
                              crossAxisSpacing: 12,
                              mainAxisSpacing: 12,
                            ),
                            itemCount: _products.length > 4 ? 4 : _products.length,
                            itemBuilder: (context, index) {
                              return ProductCard(product: _products[index]);
                            },
                          ),
                  ],
                ),
              ),

              // Trending
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Trending Now',
                              style: GoogleFonts.playfairDisplay(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const Text(
                              'Paling diminati',
                              style: TextStyle(color: Colors.grey),
                            ),
                          ],
                        ),
                        TextButton(
                          onPressed: () {},
                          child: const Text('Lihat Semua'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _isLoading
                        ? const Center(child: CircularProgressIndicator())
                        : GridView.builder(
                            shrinkWrap: true,
                            physics: const NeverScrollableScrollPhysics(),
                            gridDelegate:
                                const SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: 2,
                              childAspectRatio: 0.65,
                              crossAxisSpacing: 12,
                              mainAxisSpacing: 12,
                            ),
                            itemCount: _products.length > 8 ? 4 : (_products.length - 4).clamp(0, 4),
                            itemBuilder: (context, index) {
                              return ProductCard(product: _products[index + 4]);
                            },
                          ),
                  ],
                ),
              ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
