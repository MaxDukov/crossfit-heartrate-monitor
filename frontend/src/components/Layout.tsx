import { NavLink, Outlet } from "react-router-dom";
import { useHrStore } from "../lib/store";
import { useEffect } from "react";

const navItems = [
  { to: "/", label: "Монитор" },
  { to: "/athletes", label: "Спортсмены" },
  { to: "/sensors", label: "Датчики" },
  { to: "/sessions", label: "Тренировки" },
];

export default function Layout() {
  const { connectWs, fetchSensors } = useHrStore();

  useEffect(() => {
    connectWs();
    fetchSensors();
    return () => useHrStore.getState().disconnectWs();
  }, []);

  return (
    <div className="h-screen bg-slate-950 text-slate-100 flex flex-col overflow-hidden">
      <nav className="bg-slate-900 border-b border-slate-800 px-6 py-3 flex items-center gap-8 shrink-0">
        <span className="text-xl font-bold tracking-tight text-white">
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
                    ? "bg-slate-700 text-white"
                    : "text-slate-400 hover:text-white hover:bg-slate-800"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </div>
      </nav>
      <main className="flex-1 overflow-hidden min-h-0">
        <Outlet />
      </main>
    </div>
  );
}
