import axios from "axios";
import { BiteshipArea } from "@/types/shipping";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api",
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true, // Required to send/receive cookies for session management
});

// Add auth token to requests if available
api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("auth_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Handle session expired responses - auto logout
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (typeof window !== "undefined") {
      const errorCode = error.response?.data?.error;
      
      // If session expired or unauthorized due to deleted user
      if (errorCode === "session_expired" || 
          (error.response?.status === 401 && errorCode === "unauthorized")) {
        // Clear auth data
        localStorage.removeItem("auth_token");
        localStorage.removeItem("user");
        
        // Redirect to login if not already there
        if (!window.location.pathname.includes("/login")) {
          window.location.href = "/login?session_expired=true";
        }
      }
    }
    return Promise.reject(error);
  }
);

// ============================================
// BITESHIP AREA SEARCH
// ============================================

/**
 * Search for areas using Biteship API
 * @param query - Search query (e.g., "Semarang", "50191")
 * @returns List of matching areas with area_id, name, and postal_code
 */
export async function searchAreas(query: string): Promise<BiteshipArea[]> {
  const response = await api.get<{ areas: BiteshipArea[] }>("/shipping/areas", {
    params: { q: query },
  });
  return response.data.areas || [];
}

/**
 * Get shipping rates using area_id (Biteship)
 * @param originAreaId - Origin area ID
 * @param destinationAreaId - Destination area ID
 * @param weight - Weight in grams
 */
export async function getShippingRatesByAreaId(
  originAreaId: string,
  destinationAreaId: string,
  weight: number
) {
  const response = await api.post("/shipping/rates", {
    origin_area_id: originAreaId,
    destination_area_id: destinationAreaId,
    weight,
  });
  return response.data;
}

export default api;
