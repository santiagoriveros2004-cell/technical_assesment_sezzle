import styles from "./Key.module.css";

export type KeyTone = "digit" | "function" | "operator" | "equals";
export type KeySpan = "wide" | "tall";

interface KeyProps {
  label: string;
  name: string;
  tone: KeyTone;
  span?: KeySpan;
  disabled: boolean;
  onPress: () => void;
}

export function Key({ label, name, tone, span, disabled, onPress }: KeyProps) {
  const className = [styles.key, styles[tone], span && styles[span]].filter(Boolean).join(" ");

  return (
    <button
      type="button"
      className={className}
      aria-label={name === label ? undefined : name}
      disabled={disabled}
      onClick={onPress}
    >
      {label}
    </button>
  );
}
