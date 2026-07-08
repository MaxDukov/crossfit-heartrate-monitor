import { NavLink, Outlet } from "react-router-dom";
import { useHrStore } from "../lib/store";
import { useTheme } from "../lib/theme";
import { useEffect } from "react";

const navItems = [
  { to: "/", label: "Монитор" },
  { to: "/wod-builder", label: "WoD" },
  { to: "/athletes", label: "Спортсмены" },
  { to: "/sensors", label: "Датчики" },
  { to: "/sessions", label: "Тренировки" },
  { to: "/equipment", label: "Инвентарь" },
];

export default function Layout() {
  const { connectWs, fetchSensors } = useHrStore();
  const { theme, toggleTheme } = useTheme();

  useEffect(() => {
    connectWs();
    fetchSensors();
    return () => useHrStore.getState().disconnectWs();
  }, []);

  return (
    <div className="h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 flex flex-col overflow-hidden">
      <main className="flex-1 overflow-hidden min-h-0">
        <Outlet />
      </main>
      <nav className="bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800 px-6 py-3 flex items-center gap-8 shrink-0">
        <span className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
          CF-Monitor
        </span>
        <div className="flex gap-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-slate-200 dark:bg-slate-700 text-slate-900 dark:text-white"
                    : "text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </div>
        <button
          onClick={toggleTheme}
          className="ml-auto p-2 rounded-lg text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-xl"
          title={theme === "light" ? "Тёмная тема" : "Светлая тема"}
        >
          {theme === "light" ? "🌙" : "☀️"}
        </button>
      </nav>
    </div>
  );
}
