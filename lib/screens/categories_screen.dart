import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../widgets/category_card.dart';

class CategoriesScreen extends StatelessWidget {
  const CategoriesScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final categories = [
      {
        'name': 'WANITA',
        'image': 'https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=800&q=80',
      },
      {
        'name': 'PRIA',
        'image': 'https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=800&q=80',
      },
      {
        'name': 'ANAK',
        'image': 'https://images.unsplash.com/photo-1503944583220-79d8926ad5e2?w=800&q=80',
      },
      {
        'name': 'SPORTS',
        'image': 'https://images.unsplash.com/photo-1461896836934-ffe607ba8211?w=800&q=80',
      },
      {
        'name': 'LUXURY',
        'image': 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1600&q=80',
      },
      {
        'name': 'BEAUTY',
        'image': 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=800&q=80',
      },
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('Kategori'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Semua Kategori',
              style: GoogleFonts.playfairDisplay(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Pilih kategori yang kamu suka',
              style: TextStyle(color: Colors.grey),
            ),
            const SizedBox(height: 24),
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 1.3,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
              ),
              itemCount: categories.length,
              itemBuilder: (context, index) {
                final category = categories[index];
                return CategoryCard(
                  name: category['name']!,
                  image: category['image']!,
                  onTap: () {
                    // Navigate to category products
                  },
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
