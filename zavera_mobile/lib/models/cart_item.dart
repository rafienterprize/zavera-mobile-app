import 'product.dart';

class CartItem {
  final int id;
  final Product product;
  int quantity;
  final String? selectedSize;
  final int? variantId;

  CartItem({
    required this.id,
    required this.product,
    required this.quantity,
    this.selectedSize,
    this.variantId,
  });

  double get totalPrice => product.price * quantity;

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'product': {
        'id': product.id,
        'name': product.name,
        'price': product.price,
        'image_url': product.primaryImage,
      },
      'quantity': quantity,
      'selectedSize': selectedSize,
      'variant_id': variantId,
    };
  }

  factory CartItem.fromJson(Map<String, dynamic> json) {
    return CartItem(
      id: json['id'] ?? 0,
      product: Product.fromJson(json['product']),
      quantity: json['quantity'] ?? 1,
      selectedSize: json['selectedSize'] ?? json['selected_size'],
      variantId: json['variant_id'],
    );
  }
}
