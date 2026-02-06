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

  Future<Map<String, dynamic>> login(String email, String password) async {
    try {
      final result = await _apiService.login(email, password);
      if (result != null) {
        _user = User.fromJson(result['user']);
        notifyListeners();
        return {'success': true};
      }
      return {'success': false, 'message': 'Login gagal'};
    } catch (e) {
      return {'success': false, 'message': e.toString()};
    }
  }

  Future<Map<String, dynamic>> register(Map<String, dynamic> userData) async {
    try {
      final result = await _apiService.register(userData);
      if (result != null) {
        _user = User.fromJson(result['user']);
        notifyListeners();
        return {'success': true};
      }
      return {'success': false, 'message': 'Registrasi gagal'};
    } catch (e) {
      return {'success': false, 'message': e.toString()};
    }
  }

  Future<Map<String, dynamic>> loginWithGoogle(String idToken) async {
    try {
      final result = await _apiService.googleLogin(idToken);
      if (result != null) {
        _user = User.fromJson(result['user']);
        notifyListeners();
        return {'success': true};
      }
      return {'success': false, 'message': 'Google login gagal'};
    } catch (e) {
      return {'success': false, 'message': e.toString()};
    }
  }

  Future<void> logout() async {
    await _apiService.logout();
    _user = null;
    notifyListeners();
  }

  Future<void> refreshUser() async {
    _user = await _apiService.getCurrentUser();
    notifyListeners();
  }
}
