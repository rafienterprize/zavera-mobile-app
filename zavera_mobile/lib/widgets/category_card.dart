import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:google_fonts/google_fonts.dart';

class CategoryCard extends StatelessWidget {
  final String name;
  final String image;
  final VoidCallback onTap;

  const CategoryCard({
    super.key,
    required this.name,
    required this.image,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        height: 180,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.15),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: Stack(
            fit: StackFit.expand,
            children: [
              CachedNetworkImage(
                imageUrl: image,
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
                      name,
                      style: GoogleFonts.playfairDisplay(
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                        letterSpacing: 1,
                      ),
                    ),
                    const SizedBox(height: 8),
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
                          fontSize: 11,
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
    );
  }
}

class CategoryGrid extends StatelessWidget {
  const CategoryGrid({super.key});

  @override
  Widget build(BuildContext context) {
    final categories = [
      {
        'name': 'Wanita',
        'subtitle': 'Koleksi Eksklusif',
        'image': 'https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=800&q=80',
      },
      {
        'name': 'Pria',
        'subtitle': 'Gaya Maskulin',
        'image': 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=800&q=80',
      },
      {
        'name': 'Sports',
        'subtitle': 'Aktivewear Premium',
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

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 0.85,
        crossAxisSpacing: 16,
        mainAxisSpacing: 16,
      ),
      itemCount: categories.length,
      itemBuilder: (context, index) {
        final category = categories[index];
        return CategoryCard(
          name: category['name']!,
          image: category['image']!,
          onTap: () {
            // Navigate to category page
          },
        );
      },
    );
  }
}
