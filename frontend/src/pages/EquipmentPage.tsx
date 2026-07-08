import { useEffect, useState } from "react";
import { api } from "../lib/api";
import type { Equipment, GymInventoryItem } from "../types";

export default function EquipmentPage() {
  const [equipment, setEquipment] = useState<Equipment[]>([]);
  const [inventory, setInventory] = useState<Record<string, number>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    Promise.all([api.equipment.list(), api.equipment.inventory()]).then(
      ([eq, inv]) => {
        setEquipment(eq);
        const map: Record<string, number> = {};
        inv.forEach((i) => (map[i.equipment_key] = i.quantity));
        setInventory(map);
        setLoading(false);
      }
    );
  }, []);

  const toggle = (key: string) => {
    setInventory((prev) => {
      const next = { ...prev };
      if (key in next) delete next[key];
      else next[key] = 1;
      return next;
    });
  };

  const setQty = (key: string, qty: number) => {
    setInventory((prev) => ({ ...prev, [key]: Math.max(1, qty) }));
  };

  const save = async () => {
    setSaving(true);
    const items: GymInventoryItem[] = Object.entries(inventory).map(
      ([equipment_key, quantity]) => ({ equipment_key, quantity })
    );
    await api.equipment.updateInventory(items);
    setSaving(false);
  };

  if (loading)
    return (
      <div className="flex items-center justify-center h-full text-slate-400 dark:text-slate-500 text-lg">
        Загрузка...
      </div>
    );

  const categories = [...new Set(equipment.map((e) => e.category))];

  const categoryLabels: Record<string, string> = {
    barbells: "Штанги",
    freeweights: "Свободные веса",
    rig: "Риг / Турники",
    cardio: "Кардио",
    misc: "Прочее",
  };

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Инвентарь зала</h1>
            <p className="text-slate-500 dark:text-slate-400 mt-1">
              Отметьте доступное оборудование для генерации тренировок
            </p>
          </div>
          <button
            onClick={save}
            disabled={saving}
            className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 rounded-lg font-medium text-white transition-colors"
          >
            {saving ? "Сохранение..." : "Сохранить"}
          </button>
        </div>

        <div className="space-y-6">
          {categories.map((cat) => (
            <div
              key={cat}
              className="bg-slate-100 dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5"
            >
              <h2 className="text-lg font-semibold text-slate-700 dark:text-slate-200 mb-4">
                {categoryLabels[cat] || cat}
              </h2>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                {equipment
                  .filter((e) => e.category === cat)
                  .map((e) => {
                    const checked = e.key in inventory;
                    return (
                      <div
                        key={e.key}
                        className={`flex items-center gap-3 p-3 rounded-lg border transition-colors cursor-pointer ${
                          checked
                            ? "bg-blue-50 dark:bg-slate-800 border-blue-500 dark:border-blue-600"
                            : "bg-white dark:bg-slate-800/50 border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600"
                        }`}
                        onClick={() => toggle(e.key)}
                      >
                        <span className="text-2xl">{e.icon}</span>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-slate-700 dark:text-slate-200 truncate">
                            {e.name}
                          </p>
                          {checked && (
                            <input
                              type="number"
                              min={1}
                              value={inventory[e.key]}
                              onClick={(ev) => ev.stopPropagation()}
                              onChange={(ev) =>
                                setQty(e.key, parseInt(ev.target.value) || 1)
                              }
                              className="mt-1 w-16 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-200 text-xs px-2 py-0.5 rounded border border-slate-300 dark:border-slate-600"
                            />
                          )}
                        </div>
                        <div
                          className={`w-5 h-5 rounded-md border-2 flex items-center justify-center shrink-0 ${
                            checked
                              ? "bg-blue-600 border-blue-600"
                              : "border-slate-600"
                          }`}
                        >
                          {checked && (
                            <span className="text-white text-xs">✓</span>
                          )}
                        </div>
                      </div>
                    );
                  })}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
