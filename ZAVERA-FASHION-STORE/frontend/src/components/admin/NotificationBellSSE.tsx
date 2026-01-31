'use client';

import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useAdminSSE, AdminNotification, NotificationSeverity } from '@/hooks/useAdminSSE';
import Link from 'next/link';

export function NotificationBellSSE() {
  const { notifications, unreadCount, isConnected, markAsRead, markAllAsRead } = useAdminSSE();
  const [isOpen, setIsOpen] = useState(false);
  const [toasts, setToasts] = useState<AdminNotification[]>([]);

  // Listen for toast notifications
  useEffect(() => {
    const handleNotification = (event: CustomEvent<AdminNotification>) => {
      const notification = event.detail;
      setToasts(prev => [...prev, notification]);
      
      // Auto-remove toast after 5 seconds
      setTimeout(() => {
        setToasts(prev => prev.filter(n => n.id !== notification.id));
      }, 5000);
    };

    window.addEventListener('admin-notification' as any, handleNotification);
    return () => {
      window.removeEventListener('admin-notification' as any, handleNotification);
    };
  }, []);

  const getSeverityStyles = (severity: NotificationSeverity) => {
    switch (severity) {
      case 'critical':
        return {
          bg: 'bg-red-500/10',
          border: 'border-red-500/30',
          text: 'text-red-400',
          icon: 'bg-red-500',
          glow: 'shadow-red-500/20',
        };
      case 'warning':
        return {
          bg: 'bg-yellow-500/10',
          border: 'border-yellow-500/30',
          text: 'text-yellow-400',
          icon: 'bg-yellow-500',
          glow: 'shadow-yellow-500/20',
        };
      case 'info':
      default:
        return {
          bg: 'bg-blue-500/10',
          border: 'border-blue-500/30',
          text: 'text-blue-400',
          icon: 'bg-blue-500',
          glow: 'shadow-blue-500/20',
        };
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'order_created':
        return '🛍️';
      case 'payment_received':
        return '💰';
      case 'payment_expired':
        return '⏰';
      case 'shipment_update':
        return '📦';
      case 'stock_low':
        return '⚠️';
      case 'refund_request':
        return '💸';
      case 'dispute_created':
        return '⚠️';
      case 'user_registered':
        return '👤';
      case 'user_login':
        return '🔐';
      case 'system':
        return '🎉';
      default:
        return '🔔';
    }
  };

  const getNotificationLink = (notification: AdminNotification) => {
    if (notification.data?.order_code) {
      return `/admin/orders/${notification.data.order_code}`;
    }
    if (notification.data?.dispute_code) {
      return `/admin/disputes/${notification.data.dispute_code}`;
    }
    return '/admin/dashboard';
  };

  return (
    <>
      {/* Toast Notifications */}
      <div className="fixed top-4 right-4 z-[100] space-y-2 pointer-events-none">
        <AnimatePresence>
          {toasts.map((toast) => {
            const styles = getSeverityStyles(toast.severity);
            return (
              <motion.div
                key={toast.id}
                initial={{ opacity: 0, x: 100, scale: 0.8 }}
                animate={{ opacity: 1, x: 0, scale: 1 }}
                exit={{ opacity: 0, x: 100, scale: 0.8 }}
                className={`${styles.bg} ${styles.border} border backdrop-blur-xl rounded-lg p-4 shadow-2xl ${styles.glow} max-w-sm pointer-events-auto`}
              >
                <div className="flex items-start gap-3">
                  <div className="text-2xl flex-shrink-0">
                    {getNotificationIcon(toast.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className={`font-semibold ${styles.text} text-sm`}>{toast.title}</p>
                    <p className="text-gray-300 text-xs mt-1 line-clamp-2">{toast.message}</p>
                  </div>
                  <button
                    onClick={() => setToasts(prev => prev.filter(n => n.id !== toast.id))}
                    className="text-gray-400 hover:text-white transition-colors"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>

      {/* Notification Bell */}
      <div className="relative">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="relative p-2 text-gray-400 hover:text-white transition-colors group"
          title="Notifications"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          
          {/* Connection Status Indicator */}
          <div 
            className={`absolute top-1 right-1 w-2 h-2 rounded-full transition-colors ${
              isConnected ? 'bg-green-500 shadow-lg shadow-green-500/50' : 'bg-gray-500'
            }`} 
          />
          
          {/* Badge */}
          {unreadCount > 0 && (
            <motion.span
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              className="absolute -top-1 -right-1 bg-gradient-to-br from-red-500 to-red-600 text-white text-xs font-bold rounded-full min-w-[20px] h-5 flex items-center justify-center px-1 shadow-lg shadow-red-500/50"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </motion.span>
          )}
        </button>

        {/* Dropdown Panel */}
        <AnimatePresence>
          {isOpen && (
            <>
              {/* Backdrop */}
              <div 
                className="fixed inset-0 z-40" 
                onClick={() => setIsOpen(false)}
              />
              
              {/* Panel */}
              <motion.div
                initial={{ opacity: 0, y: -10, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -10, scale: 0.95 }}
                transition={{ duration: 0.15 }}
                className="absolute right-0 mt-2 w-96 bg-gray-900 border border-gray-800 rounded-xl shadow-2xl z-50 overflow-hidden backdrop-blur-xl"
              >
                {/* Header */}
                <div className="p-4 border-b border-gray-800 flex items-center justify-between bg-gradient-to-r from-gray-900 to-gray-800">
                  <div>
                    <h3 className="font-semibold text-white">Notifications</h3>
                    <p className="text-xs text-gray-400 mt-0.5 flex items-center gap-1">
                      <span className={`inline-block w-1.5 h-1.5 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-500'}`} />
                      {isConnected ? 'Live' : 'Disconnected'}
                    </p>
                  </div>
                  {unreadCount > 0 && (
                    <button
                      onClick={markAllAsRead}
                      className="text-sm text-blue-400 hover:text-blue-300 font-medium transition-colors"
                    >
                      Mark all read
                    </button>
                  )}
                </div>

                {/* Notifications List */}
                <div className="max-h-[500px] overflow-y-auto custom-scrollbar">
                  {notifications.length === 0 ? (
                    <div className="p-8 text-center text-gray-500">
                      <svg className="w-12 h-12 mx-auto mb-2 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                      </svg>
                      <p className="text-sm">No notifications yet</p>
                      <p className="text-xs text-gray-600 mt-1">You&apos;ll be notified when something happens</p>
                    </div>
                  ) : (
                    notifications.map((notif) => {
                      const styles = getSeverityStyles(notif.severity);
                      return (
                        <Link
                          key={notif.id}
                          href={getNotificationLink(notif)}
                          onClick={() => {
                            markAsRead(notif.id);
                            setIsOpen(false);
                          }}
                          className={`block p-4 border-b border-gray-800 hover:bg-gray-800/50 cursor-pointer transition-all ${
                            !notif.read ? `${styles.bg} border-l-4 ${styles.border}` : ''
                          }`}
                        >
                          <div className="flex items-start gap-3">
                            <div className="text-2xl flex-shrink-0">
                              {getNotificationIcon(notif.type)}
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="font-medium text-white text-sm">{notif.title}</p>
                              <p className="text-sm text-gray-400 mt-1 line-clamp-2">{notif.message}</p>
                              <div className="flex items-center gap-2 mt-2">
                                <p className="text-xs text-gray-500">
                                  {formatTimestamp(notif.timestamp)}
                                </p>
                                {notif.severity !== 'info' && (
                                  <span className={`text-xs px-2 py-0.5 rounded-full ${styles.bg} ${styles.text} border ${styles.border}`}>
                                    {notif.severity}
                                  </span>
                                )}
                              </div>
                            </div>
                            {!notif.read && (
                              <div className={`w-2 h-2 ${styles.icon} rounded-full mt-2 flex-shrink-0 shadow-lg ${styles.glow}`}></div>
                            )}
                          </div>
                        </Link>
                      );
                    })
                  )}
                </div>

                {/* Footer */}
                {notifications.length > 0 && (
                  <div className="p-3 border-t border-gray-800 text-center bg-gray-900">
                    <Link 
                      href="/admin/dashboard"
                      onClick={() => setIsOpen(false)}
                      className="text-sm text-blue-400 hover:text-blue-300 font-medium transition-colors"
                    >
                      View Dashboard
                    </Link>
                  </div>
                )}
              </motion.div>
            </>
          )}
        </AnimatePresence>
      </div>

      <style jsx global>{`
        .custom-scrollbar::-webkit-scrollbar {
          width: 6px;
        }
        .custom-scrollbar::-webkit-scrollbar-track {
          background: rgba(31, 41, 55, 0.5);
        }
        .custom-scrollbar::-webkit-scrollbar-thumb {
          background: rgba(75, 85, 99, 0.8);
          border-radius: 3px;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover {
          background: rgba(107, 114, 128, 1);
        }
      `}</style>
    </>
  );
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  
  return date.toLocaleDateString('id-ID', { 
    day: 'numeric', 
    month: 'short',
    hour: '2-digit',
    minute: '2-digit'
  });
}
