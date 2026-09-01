import { useEffect, useState } from "react";
import { api } from "./api";
import type { Wod } from "../types";

function localDate(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(
    now.getDate()
  ).padStart(2, "0")}`;
}

// useActiveWod — активный WoD или тренировка, запланированная слотом
// на сегодня (дата из часов клиента). Опрос каждые 5 секунд.
export function useActiveWod(): Wod | null {
  const [wod, setWod] = useState<Wod | null>(null);

  useEffect(() => {
    let alive = true;
    const fetchActive = () =>
      api.wods.active(localDate()).then((w) => {
        if (alive) setWod(w);
      });
    fetchActive();
    const id = setInterval(fetchActive, 5000);
    return () => {
      alive = false;
      clearInterval(id);
    };
  }, []);

  return wod;
}

export { localDate };
