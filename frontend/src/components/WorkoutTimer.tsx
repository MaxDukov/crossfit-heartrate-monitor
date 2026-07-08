import { useEffect, useRef, useState } from "react";

interface WorkoutTimerProps {
  format: string;
  durationMin: number;
}

function formatTime(sec: number): string {
  const m = Math.floor(Math.abs(sec) / 60);
  const s = Math.abs(sec) % 60;
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}

export default function WorkoutTimer({ format, durationMin }: WorkoutTimerProps) {
  const isCountdown = ["amrap", "chipper", "tabata"].includes(format);
  const totalSec = durationMin * 60;

  const [elapsed, setElapsed] = useState(0);
  const [running, setRunning] = useState(false);
  const intervalRef = useRef<number | null>(null);

  useEffect(() => {
    if (running) {
      intervalRef.current = window.setInterval(() => {
        setElapsed((e) => {
          if (isCountdown && e + 1 >= totalSec) {
            setRunning(false);
            return totalSec;
          }
          return e + 1;
        });
      }, 1000);
    } else if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [running, isCountdown, totalSec]);

  const reset = () => {
    setRunning(false);
    setElapsed(0);
  };

  const displaySec = isCountdown ? totalSec - elapsed : elapsed;
  const isFinished = isCountdown && elapsed >= totalSec;

  // EMOM minute indicator
  const emomMinute = format === "emom" ? Math.floor(elapsed / 60) + 1 : 0;

  // Tabata round indicator (20s work / 10s rest)
  const tabataRound =
    format === "tabata"
      ? Math.floor(elapsed / 30) + 1
      : 0;
  const tabataPhase =
    format === "tabata" ? (elapsed % 30 < 20 ? "work" : "rest") : "";

  const timerColor = isFinished
    ? "text-red-500"
    : !running && elapsed === 0
    ? "text-slate-300"
    : isCountdown && totalSec - elapsed <= 10
    ? "text-amber-400"
    : "text-white";

  return (
    <div className="flex items-center gap-4">
      <div className="flex flex-col items-center">
        {/* Метки для Tabata/EMOM */}
        <div className="flex items-center gap-2 h-5">
          {tabataRound > 0 && tabataRound <= 8 && (
            <>
              <span className="text-xs text-slate-400">
                Раунд {tabataRound}/8
              </span>
              <span
                className={`text-xs font-bold px-2 rounded ${
                  tabataPhase === "work"
                    ? "bg-green-600/30 text-green-400"
                    : "bg-blue-600/30 text-blue-400"
                }`}
              >
                {tabataPhase === "work" ? "Работа" : "Отдых"}
              </span>
            </>
          )}
          {emomMinute > 0 && emomMinute <= durationMin && (
            <span className="text-xs text-slate-400">
              Минута {emomMinute}/{durationMin}
            </span>
          )}
        </div>
        <div className={`font-mono font-bold tabular-nums leading-none ${timerColor}`}
             style={{ fontSize: "2.5rem" }}>
          {formatTime(displaySec)}
        </div>
      </div>

      <div className="flex gap-2">
        <button
          onClick={() => setRunning(!running)}
          disabled={isFinished}
          className={`px-4 py-2 rounded-lg font-medium text-sm transition-colors ${
            running
              ? "bg-amber-600 hover:bg-amber-500 text-white"
              : "bg-green-600 hover:bg-green-500 text-white disabled:opacity-50"
          }`}
        >
          {running ? "Пауза" : isFinished ? "Готово" : "Старт"}
        </button>
        <button
          onClick={reset}
          className="px-3 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg font-medium text-sm text-slate-200 transition-colors"
        >
          Сброс
        </button>
      </div>
    </div>
  );
}
