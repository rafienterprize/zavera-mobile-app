import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/cart_item.dart';
import '../models/product.dart';

class CartProvider with ChangeNotifier {
  List<CartItem> _items = [];

  List<CartItem> get items => _items;

  int get totalItems => _items.fold(0, (sum, item) => sum + item.quantity);

  double get totalPrice => _items.fold(0, (sum, item) => sum + item.totalPrice);

  CartProvider() {
    _loadCart();
  }

  Future<void> _loadCart() async {
    final prefs = await SharedPreferences.getInstance();
    final cartData = prefs.getString('cart');
    if (cartData != null) {
      final List<dynamic> decoded = json.decode(cartData);
      _items = decoded.map((item) => CartItem.fromJson(item)).toList();
      notifyListeners();
    }
  }

  Future<void> _saveCart() async {
    final prefs = await SharedPreferences.getInstance();
    final encoded = json.encode(_items.map((item) => item.toJson()).toList());
    await prefs.setString('cart', encoded);
  }

  void addToCart(Product product, int quantity, {String? selectedSize}) {
    final existingIndex = _items.indexWhere(
      (item) => item.product.id == product.id && item.selectedSize == selectedSize,
    );

    if (existingIndex >= 0) {
      _items[existingIndex].quantity += quantity;
    } else {
      _items.add(CartItem(
        product: product,
        quantity: quantity,
        selectedSize: selectedSize,
      ));
    }

    _saveCart();
    notifyListeners();
  }

  void removeFromCart(int productId, {String? selectedSize}) {
    _items.removeWhere(
      (item) => item.product.id == productId && item.selectedSize == selectedSize,
    );
    _saveCart();
    notifyListeners();
  }

  void updateQuantity(int productId, int quantity, {String? selectedSize}) {
    final index = _items.indexWhere(
      (item) => item.product.id == productId && item.selectedSize == selectedSize,
    );

    if (index >= 0) {
      if (quantity <= 0) {
        _items.removeAt(index);
      } else {
        _items[index].quantity = quantity;
      }
      _saveCart();
      notifyListeners();
    }
  }

  void clearCart() {
    _items.clear();
    _saveCart();
    notifyListeners();
  }
}
