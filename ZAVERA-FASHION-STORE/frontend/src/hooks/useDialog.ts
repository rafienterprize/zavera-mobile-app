import { useState } from 'react';

interface ConfirmConfig {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info';
}

interface AlertConfig {
  title: string;
  message: string;
  buttonText?: string;
  variant?: 'success' | 'error' | 'info' | 'warning';
}

export function useDialog() {
  const [showConfirm, setShowConfirm] = useState(false);
  const [showAlert, setShowAlert] = useState(false);
  const [confirmConfig, setConfirmConfig] = useState<ConfirmConfig & { onConfirm: () => void }>({
    title: '',
    message: '',
    onConfirm: () => {},
  });
  const [alertConfig, setAlertConfig] = useState<AlertConfig>({
    title: '',
    message: '',
  });

  const confirm = (config: ConfirmConfig): Promise<boolean> => {
    return new Promise((resolve) => {
      setConfirmConfig({
        ...config,
        onConfirm: () => {
          setShowConfirm(false);
          resolve(true);
        },
      });
      setShowConfirm(true);
      
      // Store reject handler
      const originalOnCancel = () => {
        setShowConfirm(false);
        resolve(false);
      };
      (setConfirmConfig as any).onCancel = originalOnCancel;
    });
  };

  const alert = (config: AlertConfig): Promise<void> => {
    return new Promise((resolve) => {
      setAlertConfig(config);
      setShowAlert(true);
      
      // Auto-resolve when closed
      const originalOnClose = () => {
        setShowAlert(false);
        resolve();
      };
      (setAlertConfig as any).onClose = originalOnClose;
    });
  };

  const closeConfirm = () => {
    setShowConfirm(false);
    if ((setConfirmConfig as any).onCancel) {
      (setConfirmConfig as any).onCancel();
    }
  };

  const closeAlert = () => {
    setShowAlert(false);
    if ((setAlertConfig as any).onClose) {
      (setAlertConfig as any).onClose();
    }
  };

  return {
    // State
    showConfirm,
    showAlert,
    confirmConfig,
    alertConfig,
    
    // Methods
    confirm,
    alert,
    closeConfirm,
    closeAlert,
  };
}
