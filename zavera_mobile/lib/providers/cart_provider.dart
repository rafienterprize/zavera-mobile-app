import 'package:flutter/material.dart';
import '../models/cart_item.dart';
import '../services/api_service.dart';

class CartProvider with ChangeNotifier {
  List<CartItem> _items = [];
  bool _isLoading = false;
  final ApiService _apiService = ApiService();

  List<CartItem> get items => _items;
  bool get isLoading => _isLoading;
  int get totalItems => _items.fold(0, (sum, item) => sum + item.quantity);
  double get totalPrice => _items.fold(0, (sum, item) => sum + (item.product.price * item.quantity));

  Future<void> loadCart() async {
    _isLoading = true;
    notifyListeners();
    
    try {
      _items = await _apiService.getCart();
    } catch (e) {
      print('Error loading cart: $e');
    }
    
    _isLoading = false;
    notifyListeners();
  }

  Future<bool> addToCart(int productId, int quantity, {int? variantId}) async {
    try {
      final success = await _apiService.addToCart(productId, quantity, variantId: variantId);
      if (success) {
        await loadCart();
        return true;
      }
      return false;
    } catch (e) {
      print('Error adding to cart: $e');
      return false;
    }
  }

  Future<bool> updateQuantity(int itemId, int quantity) async {
    try {
      final success = await _apiService.updateCartItem(itemId, quantity);
      if (success) {
        await loadCart();
        return true;
      }
      return false;
    } catch (e) {
      print('Error updating quantity: $e');
      return false;
    }
  }

  Future<bool> removeFromCart(int itemId) async {
    try {
      final success = await _apiService.removeFromCart(itemId);
      if (success) {
        await loadCart();
        return true;
      }
      return false;
    } catch (e) {
      print('Error removing from cart: $e');
      return false;
    }
  }

  Future<bool> clearCart() async {
    try {
      final success = await _apiService.clearCart();
      if (success) {
        _items.clear();
        notifyListeners();
        return true;
      }
      return false;
    } catch (e) {
      print('Error clearing cart: $e');
      return false;
    }
  }

  Future<Map<String, dynamic>?> validateCart() async {
    try {
      return await _apiService.validateCart();
    } catch (e) {
      print('Error validating cart: $e');
      return null;
    }
  }
}
