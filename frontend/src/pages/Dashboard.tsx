import { useStyle } from "../lib/style";
import BasicMonitor from "../components/monitor/BasicMonitor";
import TriColumnMonitor from "../components/monitor/TriColumnMonitor";
import SplitScreenMonitor from "../components/monitor/SplitScreenMonitor";

export default function Dashboard() {
  const style = useStyle((s) => s.style);

  if (style === "tri-column") return <TriColumnMonitor />;
  if (style === "split-screen") return <SplitScreenMonitor />;
  return <BasicMonitor />;
}
