import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../models/product.dart';
import '../models/user.dart';
import '../models/cart_item.dart';

class ApiService {
  // IMPORTANT: Update this with your backend IP address
  // For local testing: http://YOUR_LAPTOP_IP:8080/api
  // For production: https://your-domain.com/api
  static const String baseUrl = 'http://localhost:8080/api';
  
  // Get auth token
  Future<String?> _getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('auth_token');
  }

  // Get headers with auth
  Future<Map<String, String>> _getHeaders() async {
    final token = await _getToken();
    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  // ==================== AUTH ====================
  
  Future<Map<String, dynamic>?> login(String email, String password) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/auth/login'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({'email': email, 'password': password}),
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('auth_token', data['token']);
        return data;
      }
      return null;
    } catch (e) {
      print('Error logging in: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> googleLogin(String idToken) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/auth/google'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({'id_token': idToken}),
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('auth_token', data['token']);
        return data;
      }
      return null;
    } catch (e) {
      print('Error with Google login: $e');
      return null;
    }
  }

  Future<bool> verifyEmail(String token) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/auth/verify-email?token=$token'),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error verifying email: $e');
      return false;
    }
  }

  Future<bool> resendVerification(String email) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/auth/resend-verification'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({'email': email}),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error resending verification: $e');
      return false;
    }
  }

  Future<Map<String, dynamic>?> register(Map<String, dynamic> userData) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/auth/register'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode(userData),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = json.decode(response.body);
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('auth_token', data['token']);
        return data;
      }
      return null;
    } catch (e) {
      print('Error registering: $e');
      return null;
    }
  }

  Future<User?> getCurrentUser() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/auth/me'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return User.fromJson(json.decode(response.body));
      }
      return null;
    } catch (e) {
      print('Error getting current user: $e');
      return null;
    }
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('auth_token');
  }

  // ==================== PRODUCTS ====================
  
  Future<List<Product>> getProducts({String? category, String? search}) async {
    try {
      var uri = Uri.parse('$baseUrl/products');
      final queryParams = <String, String>{};
      if (category != null) queryParams['category'] = category;
      if (search != null) queryParams['search'] = search;
      
      if (queryParams.isNotEmpty) {
        uri = uri.replace(queryParameters: queryParams);
      }

      final response = await http.get(uri, headers: await _getHeaders());

      if (response.statusCode == 200) {
        final List<dynamic> data = json.decode(response.body);
        return data.map((json) => Product.fromJson(json)).toList();
      }
      return [];
    } catch (e) {
      print('Error fetching products: $e');
      return [];
    }
  }

  Future<Product?> getProduct(int id) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/products/$id'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return Product.fromJson(json.decode(response.body));
      }
      return null;
    } catch (e) {
      print('Error fetching product: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getProductWithVariants(int productId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/products/$productId/with-variants'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching product with variants: $e');
      return null;
    }
  }

  Future<List<dynamic>> getProductVariants(int productId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/products/$productId/variants'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching product variants: $e');
      return [];
    }
  }

  Future<Map<String, dynamic>?> getAvailableOptions(int productId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/products/$productId/options'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching available options: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> findVariant(int productId, Map<String, String> attributes) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/products/variants/find'),
        headers: await _getHeaders(),
        body: json.encode({
          'product_id': productId,
          'attributes': attributes,
        }),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error finding variant: $e');
      return null;
    }
  }

  // ==================== VARIANTS ====================
  
  Future<Map<String, dynamic>?> getVariant(int variantId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/variants/$variantId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching variant: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getVariantBySKU(String sku) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/variants/sku/$sku'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching variant by SKU: $e');
      return null;
    }
  }

  Future<List<dynamic>> getVariantImages(int variantId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/variants/$variantId/images'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching variant images: $e');
      return [];
    }
  }

  Future<Map<String, dynamic>?> checkVariantAvailability(List<int> variantIds) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/variants/check-availability'),
        headers: await _getHeaders(),
        body: json.encode({'variant_ids': variantIds}),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error checking variant availability: $e');
      return null;
    }
  }

  // ==================== CART ====================
  
  Future<List<CartItem>> getCart() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/cart'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = json.decode(response.body);
        return data.map((json) => CartItem.fromJson(json)).toList();
      }
      return [];
    } catch (e) {
      print('Error fetching cart: $e');
      return [];
    }
  }

  Future<bool> addToCart(int productId, int quantity, {int? variantId}) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/cart/items'),
        headers: await _getHeaders(),
        body: json.encode({
          'product_id': productId,
          'quantity': quantity,
          if (variantId != null) 'variant_id': variantId,
        }),
      );

      return response.statusCode == 200 || response.statusCode == 201;
    } catch (e) {
      print('Error adding to cart: $e');
      return false;
    }
  }

  Future<bool> updateCartItem(int itemId, int quantity) async {
    try {
      final response = await http.put(
        Uri.parse('$baseUrl/cart/items/$itemId'),
        headers: await _getHeaders(),
        body: json.encode({'quantity': quantity}),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error updating cart item: $e');
      return false;
    }
  }

  Future<bool> removeFromCart(int itemId) async {
    try {
      final response = await http.delete(
        Uri.parse('$baseUrl/cart/items/$itemId'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error removing from cart: $e');
      return false;
    }
  }

  Future<bool> clearCart() async {
    try {
      final response = await http.delete(
        Uri.parse('$baseUrl/cart'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error clearing cart: $e');
      return false;
    }
  }

  Future<Map<String, dynamic>?> validateCart() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/cart/validate'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error validating cart: $e');
      return null;
    }
  }

  // ==================== WISHLIST ====================
  
  Future<List<Product>> getWishlist() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/wishlist'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = json.decode(response.body);
        return data.map((json) => Product.fromJson(json)).toList();
      }
      return [];
    } catch (e) {
      print('Error fetching wishlist: $e');
      return [];
    }
  }

  Future<bool> addToWishlist(int productId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/wishlist'),
        headers: await _getHeaders(),
        body: json.encode({'product_id': productId}),
      );

      return response.statusCode == 200 || response.statusCode == 201;
    } catch (e) {
      print('Error adding to wishlist: $e');
      return false;
    }
  }

  Future<bool> removeFromWishlist(int productId) async {
    try {
      final response = await http.delete(
        Uri.parse('$baseUrl/wishlist/$productId'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error removing from wishlist: $e');
      return false;
    }
  }

  Future<bool> moveToCart(int productId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/wishlist/$productId/move-to-cart'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error moving to cart: $e');
      return false;
    }
  }

  // ==================== SHIPPING ====================
  
  Future<List<dynamic>> getShippingProviders() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/providers'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching shipping providers: $e');
      return [];
    }
  }

  Future<List<dynamic>> getShippingRates(Map<String, dynamic> shippingData) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/shipping/rates'),
        headers: await _getHeaders(),
        body: json.encode(shippingData),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching shipping rates: $e');
      return [];
    }
  }

  Future<List<dynamic>> searchAreas(String query) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/areas?q=$query'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error searching areas: $e');
      return [];
    }
  }

  Future<List<dynamic>> getProvinces() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/provinces'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching provinces: $e');
      return [];
    }
  }

  Future<List<dynamic>> getCities(int provinceId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/cities?province_id=$provinceId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching cities: $e');
      return [];
    }
  }

  Future<List<dynamic>> getDistricts(int cityId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/districts?city_id=$cityId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching districts: $e');
      return [];
    }
  }

  Future<List<dynamic>> getSubdistricts(int districtId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/subdistricts?district_id=$districtId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching subdistricts: $e');
      return [];
    }
  }

  Future<Map<String, dynamic>?> getCartShippingPreview() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipping/preview'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching cart shipping preview: $e');
      return null;
    }
  }

  Future<List<dynamic>> getUserAddresses() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/user/addresses'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching addresses: $e');
      return [];
    }
  }

  Future<Map<String, dynamic>?> createAddress(Map<String, dynamic> addressData) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/user/addresses'),
        headers: await _getHeaders(),
        body: json.encode(addressData),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error creating address: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getAddress(int addressId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/user/addresses/$addressId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching address: $e');
      return null;
    }
  }

  Future<bool> updateAddress(int addressId, Map<String, dynamic> addressData) async {
    try {
      final response = await http.put(
        Uri.parse('$baseUrl/user/addresses/$addressId'),
        headers: await _getHeaders(),
        body: json.encode(addressData),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error updating address: $e');
      return false;
    }
  }

  Future<bool> deleteAddress(int addressId) async {
    try {
      final response = await http.delete(
        Uri.parse('$baseUrl/user/addresses/$addressId'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error deleting address: $e');
      return false;
    }
  }

  Future<bool> setDefaultAddress(int addressId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/user/addresses/$addressId/default'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error setting default address: $e');
      return false;
    }
  }

  // ==================== CHECKOUT & ORDERS ====================
  
  Future<List<dynamic>> getShippingOptions() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/checkout/shipping-options'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching shipping options: $e');
      return [];
    }
  }

  Future<Map<String, dynamic>?> checkout(Map<String, dynamic> orderData) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/checkout/shipping'),
        headers: await _getHeaders(),
        body: json.encode(orderData),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error during checkout: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getOrder(String orderCode) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/orders/$orderCode'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching order: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getOrderById(int orderId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/orders/id/$orderId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching order by ID: $e');
      return null;
    }
  }

  Future<List<dynamic>> getUserOrders() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/user/orders'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching user orders: $e');
      return [];
    }
  }

  Future<List<dynamic>> getPendingOrders() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/pembelian/pending'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching pending orders: $e');
      return [];
    }
  }

  Future<List<dynamic>> getTransactionHistory() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/pembelian/history'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching transaction history: $e');
      return [];
    }
  }

  // ==================== PAYMENT ====================
  
  // Midtrans Snap Payment (for credit card, e-wallet, etc)
  Future<Map<String, dynamic>?> initiatePayment(String orderId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/payments/initiate'),
        headers: await _getHeaders(),
        body: json.encode({'order_id': orderId}),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error initiating payment: $e');
      return null;
    }
  }

  // Virtual Account Payment (BCA, BNI, BRI, Mandiri, Permata)
  Future<Map<String, dynamic>?> createVAPayment(String orderId, String bank) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/payments/core/create'),
        headers: await _getHeaders(),
        body: json.encode({
          'order_id': orderId,
          'bank': bank,
        }),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error creating VA payment: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getPaymentDetails(String orderId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/payments/core/$orderId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching payment details: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> checkPaymentStatus(String orderId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/payments/core/check'),
        headers: await _getHeaders(),
        body: json.encode({'order_id': orderId}),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error checking payment status: $e');
      return null;
    }
  }

  // ==================== TRACKING ====================
  
  Future<Map<String, dynamic>?> getTrackingByResi(String resi) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/tracking/$resi'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching tracking: $e');
      return null;
    }
  }

  Future<Map<String, dynamic>?> getShipment(int shipmentId) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/shipments/$shipmentId'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching shipment: $e');
      return null;
    }
  }

  Future<bool> refreshTracking(int shipmentId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/shipments/$shipmentId/refresh'),
        headers: await _getHeaders(),
      );

      return response.statusCode == 200;
    } catch (e) {
      print('Error refreshing tracking: $e');
      return false;
    }
  }

  // ==================== REFUNDS ====================
  
  Future<List<dynamic>> getOrderRefunds(String orderCode) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/customer/orders/$orderCode/refunds'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return [];
    } catch (e) {
      print('Error fetching order refunds: $e');
      return [];
    }
  }

  Future<Map<String, dynamic>?> getRefundByCode(String refundCode) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/customer/refunds/$refundCode'),
        headers: await _getHeaders(),
      );

      if (response.statusCode == 200) {
        return json.decode(response.body);
      }
      return null;
    } catch (e) {
      print('Error fetching refund: $e');
      return null;
    }
  }
}
