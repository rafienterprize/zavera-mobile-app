import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class WishlistProvider with ChangeNotifier {
  Set<int> _wishlistIds = {};

  Set<int> get wishlistIds => _wishlistIds;
  int get count => _wishlistIds.length;

  WishlistProvider() {
    _loadWishlist();
  }

  Future<void> _loadWishlist() async {
    final prefs = await SharedPreferences.getInstance();
    final data = prefs.getString('wishlist');
    if (data != null) {
      _wishlistIds = Set<int>.from(json.decode(data));
      notifyListeners();
    }
  }

  Future<void> _saveWishlist() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('wishlist', json.encode(_wishlistIds.toList()));
  }

  bool isInWishlist(int productId) {
    return _wishlistIds.contains(productId);
  }

  void toggleWishlist(int productId) {
    if (_wishlistIds.contains(productId)) {
      _wishlistIds.remove(productId);
    } else {
      _wishlistIds.add(productId);
    }
    _saveWishlist();
    notifyListeners();
  }
}
