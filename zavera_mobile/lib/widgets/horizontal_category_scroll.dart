import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'dart:async';
import '../screens/category_detail_screen.dart';

class HorizontalCategoryScroll extends StatefulWidget {
  const HorizontalCategoryScroll({super.key});

  @override
  State<HorizontalCategoryScroll> createState() => _HorizontalCategoryScrollState();
}

class _HorizontalCategoryScrollState extends State<HorizontalCategoryScroll> {
  final ScrollController _scrollController = ScrollController();
  bool _isAutoScrolling = false;
  bool _userInteracting = false;
  Timer? _resumeTimer;
  
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
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _startAutoScroll();
    });
  }

  void _startAutoScroll() async {
    if (!mounted || _userInteracting || _isAutoScrolling) return;
    
    _isAutoScrolling = true;
    
    while (mounted && !_userInteracting) {
      if (!_scrollController.hasClients) {
        await Future.delayed(const Duration(milliseconds: 100));
        continue;
      }
      
      final currentOffset = _scrollController.offset;
      final maxOffset = _scrollController.position.maxScrollExtent;
      
      // If near the end, jump back to start
      if (currentOffset >= maxOffset - 100) {
        _scrollController.jumpTo(0);
        await Future.delayed(const Duration(milliseconds: 100));
        continue;
      }
      
      // Smooth scroll forward
      await _scrollController.animateTo(
        currentOffset + 100,
        duration: const Duration(milliseconds: 2000),
        curve: Curves.linear,
      );
      
      // Small delay between animations
      await Future.delayed(const Duration(milliseconds: 10));
    }
    
    _isAutoScrolling = false;
  }

  void _onUserScrollStart() {
    _resumeTimer?.cancel();
    if (!_userInteracting) {
      setState(() => _userInteracting = true);
    }
  }

  void _onUserScrollEnd() {
    // Resume auto-scroll after 1.5 seconds of user inactivity
    _resumeTimer?.cancel();
    _resumeTimer = Timer(const Duration(milliseconds: 1500), () {
      if (mounted) {
        setState(() => _userInteracting = false);
        _startAutoScroll();
      }
    });
  }

  void _handleCardTap(String categoryName) {
    // Stop auto-scroll temporarily
    setState(() => _userInteracting = true);
    
    // Navigate to category
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => CategoryDetailScreen(
          category: categoryName,
        ),
      ),
    ).then((_) {
      // Resume auto-scroll when coming back
      if (mounted) {
        setState(() => _userInteracting = false);
        _startAutoScroll();
      }
    });
  }

  @override
  void dispose() {
    _resumeTimer?.cancel();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Create infinite list by repeating categories multiple times
    final infiniteCategories = [
      ...categories,
      ...categories,
      ...categories,
      ...categories,
      ...categories,
    ];
    
    return SizedBox(
      height: 200,
      child: GestureDetector(
        onPanDown: (_) {
          // User touched - pause auto scroll
          _onUserScrollStart();
        },
        onPanEnd: (_) {
          // User released - resume auto scroll after delay
          _onUserScrollEnd();
        },
        onPanCancel: () {
          // User cancelled - resume auto scroll after delay
          _onUserScrollEnd();
        },
        onTap: () {
          // Do nothing - let the child GestureDetector handle it
        },
        child: ListView.builder(
          controller: _scrollController,
          scrollDirection: Axis.horizontal,
          physics: const BouncingScrollPhysics(),
          itemCount: infiniteCategories.length,
          itemBuilder: (context, index) {
            final category = infiniteCategories[index];
            return Padding(
              padding: EdgeInsets.only(
                right: 16,
                left: index == 0 ? 0 : 0,
              ),
              child: GestureDetector(
                onTap: () => _handleCardTap(category['name']!),
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
      ),
    );
  }
}
