/// <reference types="react" />
/// <reference types="react-dom" />

declare module "framer-motion" {
  import * as React from "react";

  export interface AnimationProps {
    initial?: any;
    animate?: any;
    exit?: any;
    transition?: any;
    variants?: any;
    whileHover?: any;
    whileTap?: any;
    whileFocus?: any;
    whileDrag?: any;
    whileInView?: any;
    viewport?: any;
  }

  export const motion: {
    [K in keyof JSX.IntrinsicElements]: React.ForwardRefExoticComponent<
      JSX.IntrinsicElements[K] & AnimationProps
    >;
  };

  export const AnimatePresence: React.FC<{
    children?: React.ReactNode;
    mode?: "sync" | "wait" | "popLayout";
    initial?: boolean;
    onExitComplete?: () => void;
  }>;

  export function useScroll(): {
    scrollX: any;
    scrollY: any;
    scrollXProgress: any;
    scrollYProgress: any;
  };

  export function useTransform(
    value: any,
    inputRange: number[],
    outputRange: number[]
  ): any;
}
