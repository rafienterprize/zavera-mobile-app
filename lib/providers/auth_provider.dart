import 'package:flutter/material.dart';
import '../models/user.dart';
import '../services/api_service.dart';

class AuthProvider with ChangeNotifier {
  User? _user;
  bool _isLoading = true;
  final ApiService _apiService = ApiService();

  User? get user => _user;
  bool get isAuthenticated => _user != null;
  bool get isLoading => _isLoading;

  AuthProvider() {
    _checkAuth();
  }

  Future<void> _checkAuth() async {
    _user = await _apiService.getCurrentUser();
    _isLoading = false;
    notifyListeners();
  }

  Future<bool> login(String email, String password) async {
    try {
      final result = await _apiService.login(email, password);
      if (result != null) {
        _user = User.fromJson(result['user']);
        notifyListeners();
        return true;
      }
      return false;
    } catch (e) {
      print('Login error: $e');
      return false;
    }
  }

  Future<bool> register(Map<String, dynamic> userData) async {
    try {
      final result = await _apiService.register(userData);
      if (result != null) {
        _user = User.fromJson(result['user']);
        notifyListeners();
        return true;
      }
      return false;
    } catch (e) {
      print('Register error: $e');
      return false;
    }
  }

  Future<void> logout() async {
    await _apiService.logout();
    _user = null;
    notifyListeners();
  }
}
