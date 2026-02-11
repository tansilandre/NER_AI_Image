import { HTMLAttributes, forwardRef } from 'react';

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'primary' | 'success' | 'error' | 'warning';
}

export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  ({ variant = 'default', children, className = '', ...props }, ref) => {
    const variants = {
      default: 'bg-gray-100 text-gray-700',
      primary: 'bg-[var(--color-blue)] text-white',
      success: 'bg-[var(--color-green)] text-white',
      error: 'bg-[var(--color-red)] text-white',
      warning: 'bg-[var(--color-yellow)] text-black',
    };

    return (
      <span
        ref={ref}
        className={`badge inline-flex items-center px-3 py-1 text-xs font-heading font-semibold uppercase tracking-wider ${variants[variant]} ${className}`}
        {...props}
      >
        {children}
      </span>
    );
  }
);

Badge.displayName = 'Badge';
