import React, { useRef, useState } from 'react';
import { IconButton } from '@entur/button';
import { CopyIcon, CheckIcon } from '@entur/icons';

type Props = { children: React.ReactNode };

export const CopyCommand = ({ children }: Props) => {
  const [copied, setCopied] = useState(false);
  const textRef = useRef<HTMLSpanElement>(null);

  const handleCopy = () => {
    const text = textRef.current?.textContent ?? '';
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <pre className="eds-preformatted-text eds-copyable-text__preformatted-text">
      <span ref={textRef} className="eds-copyable-text__displayed-text">{children}</span>
      <IconButton
        aria-label="Copy to clipboard"
        onClick={handleCopy}
        className="eds-copyable-text__button"
        type="button"
      >
        {copied ? <CheckIcon /> : <CopyIcon />}
      </IconButton>
    </pre>
  );
};
