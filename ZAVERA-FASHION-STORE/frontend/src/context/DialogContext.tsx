'use client';

import { createContext, useContext, useState, ReactNode } from 'react';
import ConfirmDialog from '@/components/ui/ConfirmDialog';
import AlertDialog from '@/components/ui/AlertDialog';

interface ConfirmOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info';
}

interface AlertOptions {
  title: string;
  message: string;
  buttonText?: string;
  variant?: 'success' | 'error' | 'info' | 'warning';
}

interface DialogContextType {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
  alert: (options: AlertOptions) => Promise<void>;
}

const DialogContext = createContext<DialogContextType | undefined>(undefined);

export function DialogProvider({ children }: { children: ReactNode }) {
  const [showConfirm, setShowConfirm] = useState(false);
  const [showAlert, setShowAlert] = useState(false);
  const [confirmOptions, setConfirmOptions] = useState<ConfirmOptions>({
    title: '',
    message: '',
  });
  const [alertOptions, setAlertOptions] = useState<AlertOptions>({
    title: '',
    message: '',
  });
  const [confirmResolver, setConfirmResolver] = useState<((value: boolean) => void) | null>(null);
  const [alertResolver, setAlertResolver] = useState<(() => void) | null>(null);

  const confirm = (options: ConfirmOptions): Promise<boolean> => {
    return new Promise((resolve) => {
      setConfirmOptions(options);
      setConfirmResolver(() => resolve);
      setShowConfirm(true);
    });
  };

  const alert = (options: AlertOptions): Promise<void> => {
    return new Promise((resolve) => {
      setAlertOptions(options);
      setAlertResolver(() => resolve);
      setShowAlert(true);
    });
  };

  const handleConfirm = () => {
    if (confirmResolver) {
      confirmResolver(true);
      setConfirmResolver(null);
    }
    setShowConfirm(false);
  };

  const handleCancel = () => {
    if (confirmResolver) {
      confirmResolver(false);
      setConfirmResolver(null);
    }
    setShowConfirm(false);
  };

  const handleAlertClose = () => {
    if (alertResolver) {
      alertResolver();
      setAlertResolver(null);
    }
    setShowAlert(false);
  };

  return (
    <DialogContext.Provider value={{ confirm, alert }}>
      {children}
      
      <ConfirmDialog
        isOpen={showConfirm}
        title={confirmOptions.title}
        message={confirmOptions.message}
        confirmText={confirmOptions.confirmText || 'Confirm'}
        cancelText={confirmOptions.cancelText || 'Cancel'}
        variant={confirmOptions.variant || 'warning'}
        onConfirm={handleConfirm}
        onCancel={handleCancel}
      />
      
      <AlertDialog
        isOpen={showAlert}
        title={alertOptions.title}
        message={alertOptions.message}
        buttonText={alertOptions.buttonText || 'OK'}
        variant={alertOptions.variant || 'info'}
        onClose={handleAlertClose}
      />
    </DialogContext.Provider>
  );
}

export function useDialog() {
  const context = useContext(DialogContext);
  if (!context) {
    throw new Error('useDialog must be used within DialogProvider');
  }
  return context;
}
