import { useEffect, useState } from "react";
import type { Athlete } from "../types";
import { api } from "../lib/api";

export default function AthletesPage() {
  const [athletes, setAthletes] = useState<Athlete[]>([]);
  const [name, setName] = useState("");
  const [maxHr, setMaxHr] = useState(190);
  const [editing, setEditing] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editMaxHr, setEditMaxHr] = useState(190);
  const [status, setStatus] = useState<string | null>(null);

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
      await load();
    } catch (e: any) {
      setStatus("Ошибка: " + (e.message || e));
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleCreate();
    }
  };

  const handleUpdate = async (id: string) => {
    try {
      await api.athletes.update(id, { name: editName.trim(), max_hr: editMaxHr });
      setEditing(null);
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
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Спортсмены</h1>

      <div className="flex gap-3 mb-2">
        <input
          className="flex-1 bg-slate-800 border border-slate-700 rounded px-3 py-2 text-sm text-white"
          placeholder="Имя"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <input
          className="w-24 bg-slate-800 border border-slate-700 rounded px-3 py-2 text-sm text-white"
          type="number"
          placeholder="Max HR"
          value={maxHr}
          onChange={(e) => setMaxHr(Number(e.target.value))}
        />
        <button
          className="bg-blue-600 hover:bg-blue-500 px-4 py-2 rounded text-sm font-medium text-white"
          onClick={handleCreate}
        >
          Добавить
        </button>
      </div>

      {status && (
        <div className="bg-yellow-900/50 border border-yellow-700 text-yellow-200 px-4 py-2 rounded text-sm mb-4">
          {status}
        </div>
      )}

      <div className="space-y-2 mt-4">
        {athletes.map((a) =>
          editing === a.id ? (
            <div key={a.id} className="flex gap-2 bg-slate-800/50 border border-slate-700 rounded-lg p-3">
              <input
                className="flex-1 bg-slate-700 border border-slate-600 rounded px-3 py-1.5 text-sm text-white"
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
              />
              <input
                className="w-24 bg-slate-700 border border-slate-600 rounded px-3 py-1.5 text-sm text-white"
                type="number"
                value={editMaxHr}
                onChange={(e) => setEditMaxHr(Number(e.target.value))}
              />
              <button
                className="bg-emerald-600 hover:bg-emerald-500 px-3 py-1.5 rounded text-sm text-white"
                onClick={() => handleUpdate(a.id)}
              >
                Сохранить
              </button>
              <button
                className="text-slate-400 hover:text-white px-2 text-sm"
                onClick={() => setEditing(null)}
              >
                Отмена
              </button>
            </div>
          ) : (
            <div
              key={a.id}
              className="flex items-center justify-between bg-slate-800/50 border border-slate-700 rounded-lg p-3"
            >
              <div>
                <span className="font-medium">{a.name}</span>
                <span className="text-slate-400 text-sm ml-3">
                  Max HR: {a.max_hr}
                </span>
              </div>
              <div className="flex gap-2">
                <button
                  className="text-sm text-slate-400 hover:text-white px-2"
                  onClick={() => startEdit(a)}
                >
                  Ред.
                </button>
                <button
                  className="text-sm text-red-400 hover:text-red-300 px-2"
                  onClick={() => handleDelete(a.id)}
                >
                  Удалить
                </button>
              </div>
            </div>
          )
        )}
      </div>
    </div>
  );
}
