/* pages/RemovalsPage.tsx — Data removal requests and status */

import { useQuery } from "@tanstack/react-query";
import { Trash2, CheckCircle, Clock, AlertTriangle } from "lucide-react";
import { api } from "../api/client";

interface RemovalRequest {
  id: string;
  pii_type: string;
  source_domain: string;
  status: "pending" | "in_progress" | "completed" | "failed";
  created_at: string;
  completed_at: string | null;
  error_message: string | null;
}

export function RemovalsPage() {
  const { data: removals = [] } = useQuery({
    queryKey: ["removals"],
    queryFn: async () => {
      const r = await api.listRemovals();
      return r.data as RemovalRequest[];
    },
  });

  const getStatusIcon = (status: RemovalRequest["status"]) => {
    switch (status) {
      case "completed":
        return <CheckCircle size={16} className="text-emerald-400" />;
      case "in_progress":
        return <Clock size={16} className="text-blue-400 animate-pulse" />;
      case "failed":
        return <AlertTriangle size={16} className="text-red-400" />;
      default:
        return <Clock size={16} className="text-slate-400" />;
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2">
          <Trash2 size={24} />
          Removal Requests
        </h1>
        <p className="text-slate-400 text-sm mt-1">
          Track your data removal requests across the web
        </p>
      </div>

      <div className="space-y-3">
        {removals.length === 0 ? (
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-8 text-center">
            <Trash2 size={32} className="text-slate-600 mx-auto mb-3" />
            <p className="text-slate-400">No removal requests yet</p>
          </div>
        ) : (
          removals.map((req) => (
            <div
              key={req.id}
              className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center justify-between"
            >
              <div className="flex items-center gap-4 flex-1">
                <div className="flex-shrink-0">{getStatusIcon(req.status)}</div>
                <div>
                  <p className="text-sm font-medium text-slate-200">{req.source_domain}</p>
                  <p className="text-xs text-slate-500 mt-0.5">{req.pii_type}</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xs font-medium text-slate-400 capitalize">{req.status}</p>
                <p className="text-xs text-slate-600 mt-0.5">
                  {new Date(req.created_at).toLocaleDateString()}
                </p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
