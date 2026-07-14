import { useEffect, useState } from "react";
import type { Athlete } from "../types";
import { api } from "../lib/api";
import OnScreenKeyboard from "../components/OnScreenKeyboard";

type ActiveField = "name" | "maxHr" | "editName" | "editMaxHr" | null;

export default function AthletesPage() {
  const [athletes, setAthletes] = useState<Athlete[]>([]);
  const [name, setName] = useState("");
  const [maxHr, setMaxHr] = useState(190);
  const [editing, setEditing] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editMaxHr, setEditMaxHr] = useState(190);
  const [status, setStatus] = useState<string | null>(null);
  const [kbField, setKbField] = useState<ActiveField>(null);

  const load = async () => {
    try {
      const list = await api.athletes.list();
      setAthletes(list);
    } catch (e: any) {
      setStatus("Ошибка загрузки: " + (e.message || e));
    }
  };

  useEffect(() => { load(); }, []);

  const handleCreate = async () => {
    if (!name.trim()) {
      setStatus("Введите имя");
      return;
    }
    setStatus("Создание...");
    try {
      await api.athletes.create({ name: name.trim(), max_hr: maxHr });
      setName("");
      setStatus(null);
      setKbField(null);
      await load();
    } catch (e: any) {
      setStatus("Ошибка: " + (e.message || e));
    }
  };

  const handleUpdate = async (id: string) => {
    try {
      await api.athletes.update(id, { name: editName.trim(), max_hr: editMaxHr });
      setEditing(null);
      setKbField(null);
      load();
    } catch (e: any) {
      setStatus("Ошибка обновления: " + (e.message || e));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Удалить спортсмена?")) return;
    try {
      await api.athletes.delete(id);
      load();
    } catch (e: any) {
      setStatus("Ошибка удаления: " + (e.message || e));
    }
  };

  const startEdit = (a: Athlete) => {
    setEditing(a.id);
    setEditName(a.name);
    setEditMaxHr(a.max_hr);
    setKbField(null);
  };

  const insertText = (text: string, val: string, setter: (v: string) => void) => {
    setter(val + text);
  };

  const backspace = (val: string, setter: (v: string) => void) => {
    setter(val.slice(0, -1));
  };

  const onKbKey = (ch: string) => {
    switch (kbField) {
      case "name": return insertText(ch, name, setName);
      case "maxHr": return insertText(ch, String(maxHr), (v) => setMaxHr(Number(v) || 0));
      case "editName": return insertText(ch, editName, setEditName);
      case "editMaxHr": return insertText(ch, String(editMaxHr), (v) => setEditMaxHr(Number(v) || 0));
    }
  };

  const onKbBackspace = () => {
    switch (kbField) {
      case "name": return backspace(name, setName);
      case "maxHr": return backspace(String(maxHr), (v) => setMaxHr(Number(v) || 0));
      case "editName": return backspace(editName, setEditName);
      case "editMaxHr": return backspace(String(editMaxHr), (v) => setEditMaxHr(Number(v) || 0));
    }
  };

  const onKbEnter = () => {
    if (editing) {
      handleUpdate(editing);
    } else {
      handleCreate();
    }
  };

  const activeRing = (field: ActiveField) =>
    kbField === field
      ? "ring-2 ring-blue-500 ring-offset-1 dark:ring-offset-slate-900"
      : "";

  return (
    <div className="max-w-2xl mx-auto p-6" style={{ paddingBottom: kbField ? "34vh" : undefined }}>
      <h1 className="text-2xl font-bold mb-6 text-slate-900 dark:text-slate-100">Спортсмены</h1>

      <div className="flex gap-3 mb-2">
        <input
          className={`flex-1 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white ${activeRing("name")}`}
          placeholder="Имя"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); handleCreate(); } }}
          onFocus={() => setKbField("name")}
          readOnly={kbField === "name"}
        />
        <input
          className={`w-24 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white ${activeRing("maxHr")}`}
          type="number"
          placeholder="Max HR"
          value={maxHr}
          onChange={(e) => setMaxHr(Number(e.target.value))}
          onFocus={() => setKbField("maxHr")}
          readOnly={kbField === "maxHr"}
        />
        <button
          className="bg-blue-600 hover:bg-blue-500 px-4 py-2 rounded text-sm font-medium text-white"
          onClick={handleCreate}
        >
          Добавить
        </button>
      </div>

      {status && (
        <div className="bg-yellow-100 dark:bg-yellow-900/50 border border-yellow-300 dark:border-yellow-700 text-yellow-800 dark:text-yellow-200 px-4 py-2 rounded text-sm mb-4">
          {status}
        </div>
      )}

      <div className="space-y-2 mt-4">
        {athletes.map((a) =>
          editing === a.id ? (
            <div key={a.id} className="flex gap-2 bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3">
              <input
                className={`flex-1 bg-slate-100 dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded px-3 py-1.5 text-sm text-slate-900 dark:text-white ${activeRing("editName")}`}
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                onFocus={() => setKbField("editName")}
                readOnly={kbField === "editName"}
              />
              <input
                className={`w-24 bg-slate-100 dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded px-3 py-1.5 text-sm text-slate-900 dark:text-white ${activeRing("editMaxHr")}`}
                type="number"
                value={editMaxHr}
                onChange={(e) => setEditMaxHr(Number(e.target.value))}
                onFocus={() => setKbField("editMaxHr")}
                readOnly={kbField === "editMaxHr"}
              />
              <button
                className="bg-emerald-600 hover:bg-emerald-500 px-3 py-1.5 rounded text-sm text-white"
                onClick={() => handleUpdate(a.id)}
              >
                Сохранить
              </button>
              <button
                className="text-slate-400 hover:text-slate-900 dark:hover:text-white px-2 text-sm"
                onClick={() => { setEditing(null); setKbField(null); }}
              >
                Отмена
              </button>
            </div>
          ) : (
            <div
              key={a.id}
              className="flex items-center justify-between bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3"
            >
              <div>
                <span className="font-medium text-slate-900 dark:text-slate-100">{a.name}</span>
                <span className="text-slate-400 text-sm ml-3">
                  Max HR: {a.max_hr}
                </span>
              </div>
              <div className="flex gap-2">
                <button
                  className="text-sm text-slate-400 hover:text-slate-900 dark:hover:text-white px-2"
                  onClick={() => startEdit(a)}
                >
                  Ред.
                </button>
                <button
                  className="text-sm text-red-500 dark:text-red-400 hover:text-red-600 dark:hover:text-red-300 px-2"
                  onClick={() => handleDelete(a.id)}
                >
                  Удалить
                </button>
              </div>
            </div>
          )
        )}
      </div>

      {kbField && (
        <OnScreenKeyboard
          onKey={onKbKey}
          onBackspace={onKbBackspace}
          onEnter={onKbEnter}
          onClose={() => setKbField(null)}
        />
      )}
    </div>
  );
}
