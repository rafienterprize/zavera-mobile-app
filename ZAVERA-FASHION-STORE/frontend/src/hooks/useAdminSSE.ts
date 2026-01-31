import { useEffect, useState, useCallback, useRef } from 'react';

export type NotificationSeverity = 'info' | 'warning' | 'critical';

export interface AdminNotification {
  id: string;
  type: string;
  title: string;
  message: string;
  severity: NotificationSeverity;
  data?: any;
  timestamp: string;
  read: boolean;
}

export function useAdminSSE() {
  const [notifications, setNotifications] = useState<AdminNotification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isConnected, setIsConnected] = useState(false);
  const eventSourceRef = useRef<{ close: () => void } | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 10;

  const parseSSEMessage = useCallback((message: string) => {
    const lines = message.split('\n');
    let eventType = 'message';
    let data = '';

    lines.forEach(line => {
      if (line.startsWith('event:')) {
        eventType = line.substring(6).trim();
      } else if (line.startsWith('data:')) {
        data = line.substring(5).trim();
      } else if (line.startsWith(':')) {
        // Comment (keepalive)
        return;
      }
    });

    if (!data) return;

    try {
      if (eventType === 'connected') {
        const connData = JSON.parse(data);
        console.log('📢 SSE connected:', connData.message);
        return;
      }

      if (eventType === 'notification') {
        const notification: AdminNotification = JSON.parse(data);
        console.log('📢 Notification received:', notification);

        // Add to notifications list
        setNotifications(prev => [notification, ...prev]);
        setUnreadCount(prev => prev + 1);

        // Play sound
        playNotificationSound();

        // Show browser notification
        if ('Notification' in window && Notification.permission === 'granted') {
          new Notification(notification.title, {
            body: notification.message,
            icon: '/favicon.ico',
            tag: notification.id,
            badge: '/favicon.ico',
          });
        }

        // Show toast notification
        showToast(notification);
      }
    } catch (err) {
      console.error('❌ Failed to parse SSE message:', err);
    }
  }, []);

  const connect = useCallback(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      console.log('❌ No auth token, skipping SSE connection');
      return;
    }

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }

    // SSE URL
    const protocol = window.location.protocol === 'https:' ? 'https:' : 'http:';
    const host = window.location.hostname === 'localhost' ? 'localhost:8080' : window.location.host;
    const sseUrl = `${protocol}//${host}/api/admin/events`;

    console.log('🔌 Connecting to SSE...');

    const controller = new AbortController();
    let isAborted = false;
    
    fetch(sseUrl, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Accept': 'text/event-stream',
      },
      signal: controller.signal,
    })
      .then(response => {
        if (!response.ok) {
          throw new Error(`SSE connection failed: ${response.status}`);
        }
        
        if (!response.body) {
          throw new Error('Response body is null');
        }

        console.log('✅ SSE connected');
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        const processStream = (): void => {
          if (isAborted) return;
          
          reader.read().then(({ done, value }) => {
            if (isAborted) return;
            
            if (done) {
              console.log('🔌 SSE stream ended');
              setIsConnected(false);
              if (!isAborted) {
                scheduleReconnect();
              }
              return;
            }

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n\n');
            buffer = lines.pop() || '';

            lines.forEach(line => {
              if (line.trim()) {
                parseSSEMessage(line);
              }
            });

            processStream();
          }).catch(err => {
            if (isAborted) return;
            
            // Ignore abort errors
            if (err.name === 'AbortError') {
              console.log('🔌 SSE connection aborted');
              return;
            }
            
            console.error('❌ SSE read error:', err);
            setIsConnected(false);
            scheduleReconnect();
          });
        };

        processStream();
      })
      .catch(err => {
        if (isAborted) return;
        
        // Ignore abort errors
        if (err.name === 'AbortError') {
          console.log('🔌 SSE connection aborted');
          return;
        }
        
        console.error('❌ SSE connection error:', err);
        setIsConnected(false);
        scheduleReconnect();
      });

    // Store abort controller for cleanup
    eventSourceRef.current = { 
      close: () => {
        isAborted = true;
        controller.abort();
      }
    };

    function scheduleReconnect() {
      if (isAborted) return;
      
      if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
        console.log('❌ Max reconnection attempts reached');
        return;
      }

      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
      console.log(`🔄 Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current + 1})`);

      reconnectTimeoutRef.current = setTimeout(() => {
        reconnectAttemptsRef.current++;
        connect();
      }, delay);
    }
  }, [parseSSEMessage]);

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    setIsConnected(false);
  }, []);

  const markAsRead = useCallback((notificationId: string) => {
    setNotifications(prev =>
      prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, []);

  const markAllAsRead = useCallback(() => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
    setUnreadCount(0);
  }, []);

  const clearNotifications = useCallback(() => {
    setNotifications([]);
    setUnreadCount(0);
  }, []);

  useEffect(() => {
    // Request notification permission
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission();
    }

    // Delay connection slightly to avoid race conditions
    const timer = setTimeout(() => {
      connect();
    }, 100);

    return () => {
      clearTimeout(timer);
      // Don't disconnect immediately on unmount
      // Let the connection stay alive
    };
  }, []); // Empty deps - only run once on mount

  // Separate cleanup on actual unmount
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, []);

  return {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead,
    clearNotifications,
  };
}

function playNotificationSound() {
  try {
    const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);

    oscillator.frequency.value = 800;
    oscillator.type = 'sine';

    gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5);

    oscillator.start(audioContext.currentTime);
    oscillator.stop(audioContext.currentTime + 0.5);
  } catch (err) {
    console.log('Sound play failed:', err);
  }
}

function showToast(notification: AdminNotification) {
  const event = new CustomEvent('admin-notification', { detail: notification });
  window.dispatchEvent(event);
}
