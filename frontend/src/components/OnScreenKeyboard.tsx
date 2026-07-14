import { useState, useCallback } from "react";

type KeyDef = string | { label: string; action: string; wide?: number };

const RU_ROWS: KeyDef[][] = [
  ["1","2","3","4","5","6","7","8","9","0","-","="],
  ["й","ц","у","к","е","н","г","ш","щ","з","х","ъ"],
  ["ф","ы","в","а","п","р","о","л","д","ж","э"],
  ["я","ч","с","м","и","т","ь","б","ю"],
];

const EN_ROWS: KeyDef[][] = [
  ["1","2","3","4","5","6","7","8","9","0","-","="],
  ["q","w","e","r","t","y","u","i","o","p"],
  ["a","s","d","f","g","h","j","k","l"],
  ["z","x","c","v","b","n","m"],
];

interface Props {
  onKey: (key: string) => void;
  onBackspace: () => void;
  onEnter: () => void;
  onClose: () => void;
}

export default function OnScreenKeyboard({ onKey, onBackspace, onEnter, onClose }: Props) {
  const [lang, setLang] = useState<"ru" | "en">("ru");
  const [shift, setShift] = useState(false);

  const rows = lang === "ru" ? RU_ROWS : EN_ROWS;
  const toggleLang = useCallback(() => { setLang(l => l === "ru" ? "en" : "ru"); }, []);
  const toggleShift = useCallback(() => setShift(s => !s), []);

  const handleChar = (ch: string) => {
    onKey(shift ? ch.toUpperCase() : ch);
    if (shift) setShift(false);
  };

  const keyClass =
    "flex items-center justify-center rounded-lg font-semibold select-none " +
    "bg-slate-700 hover:bg-slate-600 active:bg-blue-600 text-white " +
    "transition-colors text-xl cursor-pointer";

  const fnKeyClass =
    "flex items-center justify-center rounded-lg font-semibold select-none " +
    "bg-slate-600 hover:bg-slate-500 active:bg-blue-600 text-white " +
    "transition-colors text-base cursor-pointer";

  return (
    <div className="fixed bottom-0 left-0 right-0 h-[33vh] bg-slate-900/95 dark:bg-black/90 border-t-2 border-slate-600 z-50 px-4 py-2 flex flex-col gap-1.5">
      <div className="flex justify-end">
        <button
          className="text-slate-300 hover:text-white text-sm px-2 py-0.5"
          onClick={onClose}
        >
          ✕ Закрыть
        </button>
      </div>

      {rows.map((row, ri) => (
        <div key={ri} className="flex gap-1.5 justify-center flex-1">
          {ri === 3 && (
            <button
              className={`${fnKeyClass} px-3 min-w-[3rem] ${shift ? "bg-blue-700" : ""}`}
              onClick={toggleShift}
            >
              ⇧
            </button>
          )}
          {row.map((k, ki) => {
            const ch = typeof k === "string" ? k : k.label;
            return (
              <button
                key={ki}
                className={`${keyClass} flex-1 max-w-[3.2rem]`}
                onClick={() => handleChar(ch)}
              >
                {shift ? ch.toUpperCase() : ch}
              </button>
            );
          })}
          {ri === 3 && (
            <button
              className={`${fnKeyClass} px-3 min-w-[3rem]`}
              onClick={onBackspace}
            >
              ⌫
            </button>
          )}
        </div>
      ))}

      <div className="flex gap-1.5 flex-1">
        <button
          className={`${fnKeyClass} px-4 min-w-[4rem] uppercase`}
          onClick={toggleLang}
        >
          {lang}
        </button>
        <button
          className={`${keyClass} flex-1`}
          onClick={() => onKey(" ")}
        >
          Пробел
        </button>
        <button
          className={`${fnKeyClass} px-6 min-w-[5rem]`}
          onClick={onEnter}
        >
          Enter ↵
        </button>
      </div>
    </div>
  );
}
