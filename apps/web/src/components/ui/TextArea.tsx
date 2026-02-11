import { TextareaHTMLAttributes, forwardRef } from 'react';

interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
  helpText?: string;
}

export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(
  ({ label, error, helpText, className = '', ...props }, ref) => {
    return (
      <div className="space-y-1.5">
        {label && (
          <label className="block text-sm font-medium text-gray-700">
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          className={`textarea w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none transition-colors resize-none ${
            error ? 'border-[var(--color-red)]' : ''
          } ${className}`}
          {...props}
        />
        {error && (
          <p className="text-sm text-[var(--color-red)]">{error}</p>
        )}
        {helpText && !error && (
          <p className="text-sm text-gray-500">{helpText}</p>
        )}
      </div>
    );
  }
);

TextArea.displayName = 'TextArea';
