import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:cached_network_image/cached_network_image.dart';

class CategoryDetailScreen extends StatefulWidget {
  final String category;
  final String? subcategory;

  const CategoryDetailScreen({
    super.key,
    required this.category,
    this.subcategory,
  });

  @override
  State<CategoryDetailScreen> createState() => _CategoryDetailScreenState();
}

class _CategoryDetailScreenState extends State<CategoryDetailScreen> {
  String _selectedSubcategory = 'Semua';
  final List<String> _selectedSizes = [];
  String _selectedPriceRange = 'Semua Harga';
  String _sortBy = 'Terbaru';

  final Map<String, List<String>> _subcategories = {
    'Wanita': ['Semua', 'Dress', 'Atasan', 'Bawahan', 'Outerwear', 'Aksesoris'],
    'Pria': ['Semua', 'Shirts', 'Pants', 'Jackets', 'Suits'],
    'Sports': ['Semua', 'Activewear', 'Footwear', 'Accessories'],
    'Anak': ['Semua', 'Boys', 'Girls', 'Baby'],
    'Luxury': ['Semua', 'Designer', 'Premium', 'Limited Edition'],
    'Beauty': ['Semua', 'Skincare', 'Makeup', 'Fragrance'],
  };

  final Map<String, String> _categoryImages = {
    'Wanita': 'https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=1600&q=80',
    'Pria': 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=1600&q=80',
    'Sports': 'https://images.unsplash.com/photo-1461896836934-ffe607ba8211?w=1600&q=80',
    'Anak': 'https://images.unsplash.com/photo-1503944583220-79d8926ad5e2?w=1600&q=80',
    'Luxury': 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1600&q=80',
    'Beauty': 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=1600&q=80',
  };

  final Map<String, String> _categoryDescriptions = {
    'Wanita': 'Koleksi fashion wanita yang elegan dan modern untuk setiap kesempatan',
    'Pria': 'Gaya maskulin untuk pria modern dan berkelas',
    'Sports': 'Activewear premium untuk gaya hidup aktif Anda',
    'Anak': 'Fashion stylish untuk buah hati tercinta',
    'Luxury': 'Koleksi eksklusif dari brand designer ternama',
    'Beauty': 'Produk perawatan premium untuk kecantikan Anda',
  };

  void _showFilterSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        expand: false,
        builder: (context, scrollController) {
          return SingleChildScrollView(
            controller: scrollController,
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Filter',
                        style: GoogleFonts.playfairDisplay(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close),
                        onPressed: () => Navigator.pop(context),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '0 produk',
                    style: TextStyle(color: Colors.grey[600], fontSize: 14),
                  ),
                  const SizedBox(height: 24),
                  
                  // Kategori
                  _buildFilterSection(
                    'Kategori',
                    Column(
                      children: (_subcategories[widget.category] ?? []).map((sub) {
                        return RadioListTile<String>(
                          title: Text(sub),
                          value: sub,
                          groupValue: _selectedSubcategory,
                          onChanged: (value) {
                            setState(() => _selectedSubcategory = value!);
                            Navigator.pop(context);
                          },
                          contentPadding: EdgeInsets.zero,
                          dense: true,
                        );
                      }).toList(),
                    ),
                  ),
                  
                  const Divider(height: 32),
                  
                  // Ukuran
                  _buildFilterSection(
                    'Ukuran',
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: ['XS', 'S', 'M', 'L', 'XL', 'XXL'].map((size) {
                        final isSelected = _selectedSizes.contains(size);
                        return FilterChip(
                          label: Text(size),
                          selected: isSelected,
                          onSelected: (selected) {
                            setState(() {
                              if (selected) {
                                _selectedSizes.add(size);
                              } else {
                                _selectedSizes.remove(size);
                              }
                            });
                          },
                          backgroundColor: Colors.white,
                          selectedColor: const Color(0xFF1a1a1a),
                          labelStyle: TextStyle(
                            color: isSelected ? Colors.white : Colors.black,
                          ),
                          side: BorderSide(
                            color: isSelected ? const Color(0xFF1a1a1a) : Colors.grey[300]!,
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                  
                  const Divider(height: 32),
                  
                  // Harga
                  _buildFilterSection(
                    'Harga',
                    Column(
                      children: [
                        'Semua Harga',
                        'Di bawah Rp 500.000',
                        'Rp 500.000 - Rp 500.000',
                        'Rp 500.000 - Rp 500.000',
                        'Rp 500.000 - Rp 1.000.000',
                        'Di atas Rp 1.000.000',
                      ].map((price) {
                        return RadioListTile<String>(
                          title: Text(price),
                          value: price,
                          groupValue: _selectedPriceRange,
                          onChanged: (value) {
                            setState(() => _selectedPriceRange = value!);
                          },
                          contentPadding: EdgeInsets.zero,
                          dense: true,
                        );
                      }).toList(),
                    ),
                  ),
                  
                  const SizedBox(height: 24),
                  
                  // Apply Button
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () => Navigator.pop(context),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF1a1a1a),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                      child: const Text(
                        'Terapkan Filter',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildFilterSection(String title, Widget content) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 12),
        content,
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: CustomScrollView(
        slivers: [
          // Hero Banner
          SliverAppBar(
            expandedHeight: 250,
            pinned: true,
            backgroundColor: const Color(0xFF1a1a1a),
            flexibleSpace: FlexibleSpaceBar(
              background: Stack(
                fit: StackFit.expand,
                children: [
                  CachedNetworkImage(
                    imageUrl: _categoryImages[widget.category] ?? '',
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
                    bottom: 40,
                    left: 20,
                    right: 20,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          widget.category,
                          style: GoogleFonts.playfairDisplay(
                            fontSize: 36,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _categoryDescriptions[widget.category] ?? '',
                          style: const TextStyle(
                            fontSize: 14,
                            color: Colors.white,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          
          // Filter Bar
          SliverToBoxAdapter(
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                border: Border(
                  bottom: BorderSide(color: Colors.grey[200]!),
                ),
              ),
              child: Row(
                children: [
                  Text(
                    '0 Produk',
                    style: TextStyle(
                      fontSize: 14,
                      color: Colors.grey[600],
                    ),
                  ),
                  const Spacer(),
                  // Sort Dropdown
                  InkWell(
                    onTap: () {
                      showModalBottomSheet(
                        context: context,
                        backgroundColor: Colors.white,
                        shape: const RoundedRectangleBorder(
                          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
                        ),
                        builder: (context) => Container(
                          padding: const EdgeInsets.all(24),
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'Urutkan',
                                style: GoogleFonts.playfairDisplay(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              const SizedBox(height: 16),
                              ...[
                                'Terbaru',
                                'Harga: Rendah ke Tinggi',
                                'Harga: Tinggi ke Rendah',
                                'Nama A-Z',
                              ].map((sort) => RadioListTile<String>(
                                title: Text(sort),
                                value: sort,
                                groupValue: _sortBy,
                                onChanged: (value) {
                                  setState(() => _sortBy = value!);
                                  Navigator.pop(context);
                                },
                                contentPadding: EdgeInsets.zero,
                              )),
                            ],
                          ),
                        ),
                      );
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey[300]!),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Row(
                        children: [
                          Text(
                            _sortBy,
                            style: const TextStyle(fontSize: 13),
                          ),
                          const SizedBox(width: 4),
                          Icon(Icons.arrow_drop_down, size: 20, color: Colors.grey[600]),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  // Grid/List Toggle
                  IconButton(
                    icon: const Icon(Icons.grid_view, size: 20),
                    onPressed: () {},
                    padding: const EdgeInsets.all(8),
                    constraints: const BoxConstraints(),
                  ),
                  // Filter Button
                  IconButton(
                    icon: const Icon(Icons.tune, size: 20),
                    onPressed: _showFilterSheet,
                    padding: const EdgeInsets.all(8),
                    constraints: const BoxConstraints(),
                  ),
                ],
              ),
            ),
          ),
          
          // Empty State
          SliverFillRemaining(
            child: Center(
              child: Padding(
                padding: const EdgeInsets.all(32),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.inventory_2_outlined,
                      size: 80,
                      color: Colors.grey[300],
                    ),
                    const SizedBox(height: 24),
                    Text(
                      'Koleksi Segera Hadir',
                      style: GoogleFonts.playfairDisplay(
                        fontSize: 22,
                        fontWeight: FontWeight.bold,
                        color: const Color(0xFF1a1a1a),
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'Kami sedang menyiapkan koleksi terbaik untuk kategori ini.\nJelajahi kategori lainnya atau kembali lagi nanti.',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 14,
                        color: Colors.grey[600],
                        height: 1.5,
                      ),
                    ),
                    const SizedBox(height: 32),
                    OutlinedButton(
                      onPressed: () => Navigator.pop(context),
                      style: OutlinedButton.styleFrom(
                        side: const BorderSide(color: Color(0xFF1a1a1a)),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 32,
                          vertical: 14,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                      child: const Text(
                        'Jelajahi Koleksi Lain',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF1a1a1a),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
