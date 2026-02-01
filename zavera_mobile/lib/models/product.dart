class Product {
  final int id;
  final String name;
  final double price;
  final String description;
  final String? imageUrl;
  final List<String>? images;
  final int stock;
  final double? weight;
  final String category;
  final String? subcategory;
  final String? brand;
  final String? material;
  final List<String>? availableSizes;

  Product({
    required this.id,
    required this.name,
    required this.price,
    required this.description,
    this.imageUrl,
    this.images,
    required this.stock,
    this.weight,
    required this.category,
    this.subcategory,
    this.brand,
    this.material,
    this.availableSizes,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: json['id'] ?? 0,
      name: json['name'] ?? '',
      price: (json['price'] ?? 0).toDouble(),
      description: json['description'] ?? '',
      imageUrl: json['image_url'],
      images: json['images'] != null ? List<String>.from(json['images']) : null,
      stock: json['stock'] ?? 0,
      weight: json['weight']?.toDouble(),
      category: json['category'] ?? '',
      subcategory: json['subcategory'],
      brand: json['brand'],
      material: json['material'],
      availableSizes: json['available_sizes'] != null 
          ? List<String>.from(json['available_sizes']) 
          : null,
    );
  }

  String get primaryImage {
    if (images != null && images!.isNotEmpty) {
      return images!.first;
    }
    return imageUrl ?? 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=800&q=80';
  }

  bool get isLuxury => category.toLowerCase() == 'luxury';
}
