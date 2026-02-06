import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../models/product.dart';

class WishlistProvider with ChangeNotifier {
  final ApiService _apiService = ApiService();
  List<Product> _items = [];
  bool _isLoading = false;

  List<Product> get items => _items;
  int get count => _items.length;
  bool get isLoading => _isLoading;

  Future<void> loadWishlist() async {
    _isLoading = true;
    notifyListeners();
    
    final products = await _apiService.getWishlist();
    _items = products;
    
    _isLoading = false;
    notifyListeners();
  }

  bool isInWishlist(int productId) {
    return _items.any((p) => p.id == productId);
  }

  Future<void> addToWishlist(int productId) async {
    final success = await _apiService.addToWishlist(productId);
    if (success) {
      await loadWishlist();
    }
  }

  Future<void> removeFromWishlist(int productId) async {
    final success = await _apiService.removeFromWishlist(productId);
    if (success) {
      await loadWishlist();
    }
  }

  Future<void> toggleWishlist(int productId) async {
    if (isInWishlist(productId)) {
      await removeFromWishlist(productId);
    } else {
      await addToWishlist(productId);
    }
  }

  Future<void> moveToCart(int productId) async {
    final success = await _apiService.moveToCart(productId);
    if (success) {
      await loadWishlist();
    }
  }
}
