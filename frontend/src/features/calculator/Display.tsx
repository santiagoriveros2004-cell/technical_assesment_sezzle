import type { CSSProperties } from "react";
import styles from "./Display.module.css";

interface DisplayProps {
  value: string;
  expression: string;
}

export function Display({ value, expression }: DisplayProps) {
  const sizing = { "--value-length": value.length } as CSSProperties;

  return (
    <div className={styles.display}>
      <p className={styles.expression}>{expression}</p>
      <output className={styles.value} style={sizing}>
        {value}
      </output>
    </div>
  );
}
