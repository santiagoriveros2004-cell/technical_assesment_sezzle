import styles from "./App.module.css";
import { Calculator } from "./features/calculator/Calculator";

export function App() {
  return (
    <main className={styles.page}>
      <h1 className={styles.title}>Calculator</h1>
      <Calculator />
    </main>
  );
}
