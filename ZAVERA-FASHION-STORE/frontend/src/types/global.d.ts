// Global type declarations

declare global {
  interface Window {
    snap?: {
      pay: (token: string, options: {
        onSuccess: () => void;
        onPending: () => void;
        onError: (result?: unknown) => void;
        onClose: () => void;
      }) => void;
      hide: () => void;
    };
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: { credential: string }) => void;
          }) => void;
          renderButton: (
            element: HTMLElement,
            config: { theme: string; size: string; width: number }
          ) => void;
        };
      };
    };
  }
}

export {};
