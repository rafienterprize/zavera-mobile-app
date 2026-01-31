"use client";

import { X, AlertCircle, CheckCircle, Info, AlertTriangle } from "lucide-react";
import { useEffect } from "react";

interface DialogProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  message: string;
  variant?: 'success' | 'error' | 'info' | 'warning';
  buttonText?: string;
}

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'warning' | 'info';
}

export function AlertDialog({ 
  isOpen, 
  onClose, 
  title, 
  message, 
  variant = 'info',
  buttonText = 'OK'
}: DialogProps) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const variantStyles = {
    success: {
      icon: CheckCircle,
      iconColor: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/20',
      buttonColor: 'bg-emerald-500 hover:bg-emerald-600',
    },
    error: {
      icon: AlertCircle,
      iconColor: 'text-red-400',
      bgColor: 'bg-red-500/10',
      borderColor: 'border-red-500/20',
      buttonColor: 'bg-red-500 hover:bg-red-600',
    },
    warning: {
      icon: AlertTriangle,
      iconColor: 'text-yellow-400',
      bgColor: 'bg-yellow-500/10',
      borderColor: 'border-yellow-500/20',
      buttonColor: 'bg-yellow-500 hover:bg-yellow-600',
    },
    info: {
      icon: Info,
      iconColor: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      borderColor: 'border-blue-500/20',
      buttonColor: 'bg-blue-500 hover:bg-blue-600',
    },
  };

  const style = variantStyles[variant];
  const Icon = style.icon;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/60 backdrop-blur-sm animate-fade-in"
        style={{ animationDuration: '0.2s' }}
        onClick={onClose}
      />
      
      {/* Dialog */}
      <div className="relative bg-neutral-900 rounded-2xl border border-white/10 shadow-2xl max-w-md w-full animate-dialog-in">
        {/* Header */}
        <div className="flex items-start justify-between p-6 pb-4">
          <div className="flex items-start gap-4">
            <div className={`p-3 rounded-xl ${style.bgColor} border ${style.borderColor}`}>
              <Icon className={`${style.iconColor}`} size={24} />
            </div>
            <div>
              <h3 className="text-xl font-semibold text-white">{title}</h3>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-white/5 text-white/60 hover:text-white transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="px-6 pb-6">
          <p className="text-white/80 leading-relaxed whitespace-pre-line">
            {message}
          </p>
        </div>

        {/* Footer */}
        <div className="px-6 pb-6 flex justify-end">
          <button
            onClick={onClose}
            className={`px-6 py-2.5 rounded-xl text-white font-medium transition-colors ${style.buttonColor}`}
          >
            {buttonText}
          </button>
        </div>
      </div>
    </div>
  );
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  variant = 'info',
}: ConfirmDialogProps) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const variantStyles = {
    danger: {
      icon: AlertCircle,
      iconColor: 'text-red-400',
      bgColor: 'bg-red-500/10',
      borderColor: 'border-red-500/20',
      buttonColor: 'bg-red-500 hover:bg-red-600',
    },
    warning: {
      icon: AlertTriangle,
      iconColor: 'text-yellow-400',
      bgColor: 'bg-yellow-500/10',
      borderColor: 'border-yellow-500/20',
      buttonColor: 'bg-yellow-500 hover:bg-yellow-600',
    },
    info: {
      icon: Info,
      iconColor: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      borderColor: 'border-blue-500/20',
      buttonColor: 'bg-blue-500 hover:bg-blue-600',
    },
  };

  const style = variantStyles[variant];
  const Icon = style.icon;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/60 backdrop-blur-sm animate-fade-in"
        style={{ animationDuration: '0.2s' }}
        onClick={onClose}
      />
      
      {/* Dialog */}
      <div className="relative bg-neutral-900 rounded-2xl border border-white/10 shadow-2xl max-w-md w-full animate-dialog-in">
        {/* Header */}
        <div className="flex items-start justify-between p-6 pb-4">
          <div className="flex items-start gap-4">
            <div className={`p-3 rounded-xl ${style.bgColor} border ${style.borderColor}`}>
              <Icon className={`${style.iconColor}`} size={24} />
            </div>
            <div>
              <h3 className="text-xl font-semibold text-white">{title}</h3>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-white/5 text-white/60 hover:text-white transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="px-6 pb-6">
          <p className="text-white/80 leading-relaxed whitespace-pre-line">
            {message}
          </p>
        </div>

        {/* Footer */}
        <div className="px-6 pb-6 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-6 py-2.5 rounded-xl bg-white/10 text-white font-medium hover:bg-white/20 transition-colors"
          >
            {cancelText}
          </button>
          <button
            onClick={() => {
              onConfirm();
              onClose();
            }}
            className={`px-6 py-2.5 rounded-xl text-white font-medium transition-colors ${style.buttonColor}`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}

