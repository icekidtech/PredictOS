import { useEffect, useState } from "react";
import { eventApi } from "../services/api";

export default function Events() {
  const [events, setEvents] = useState<unknown[]>([]);
  useEffect(() => { eventApi.list().then((r) => setEvents(r.data.events ?? [])).catch(() => {}); }, []);
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Markets</h1>
      <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-6">
        {events.length === 0 ? <p className="text-sm text-zinc-500">No events yet. Data comes from Somnia/DreamDEX.</p> :
          <pre className="text-xs text-zinc-400 overflow-auto">{JSON.stringify(events, null, 2)}</pre>
        }
      </div>
    </div>
  );
}
