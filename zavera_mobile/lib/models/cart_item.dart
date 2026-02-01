import 'product.dart';

class CartItem {
  final Product product;
  int quantity;
  final String? selectedSize;

  CartItem({
    required this.product,
    required this.quantity,
    this.selectedSize,
  });

  double get totalPrice => product.price * quantity;

  Map<String, dynamic> toJson() {
    return {
      'product': {
        'id': product.id,
        'name': product.name,
        'price': product.price,
        'image_url': product.primaryImage,
      },
      'quantity': quantity,
      'selectedSize': selectedSize,
    };
  }

  factory CartItem.fromJson(Map<String, dynamic> json) {
    return CartItem(
      product: Product.fromJson(json['product']),
      quantity: json['quantity'] ?? 1,
      selectedSize: json['selectedSize'],
    );
  }
}
