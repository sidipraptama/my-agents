"use client";

import { Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";

export function DeleteAgentButton({ agentId }: { agentId: string }) {
  const router = useRouter();
  const [confirming, setConfirming] = useState(false);
  const [deleting, setDeleting] = useState(false);

  async function handleDelete() {
    setDeleting(true);
    try {
      await api.deleteAgent(agentId);
      router.push("/agents");
    } finally {
      setDeleting(false);
    }
  }

  if (!confirming) {
    return (
      <Button variant="ghost" className="w-full" onClick={() => setConfirming(true)}>
        <Trash2 size={14} />
        Delete Agent
      </Button>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      <p className="text-xs text-muted">This permanently deletes the agent and its tasks.</p>
      <div className="flex gap-2">
        <Button variant="secondary" className="flex-1" onClick={() => setConfirming(false)}>
          Cancel
        </Button>
        <Button variant="danger" className="flex-1" onClick={handleDelete} disabled={deleting}>
          {deleting ? "Deleting…" : "Confirm"}
        </Button>
      </div>
    </div>
  );
}
