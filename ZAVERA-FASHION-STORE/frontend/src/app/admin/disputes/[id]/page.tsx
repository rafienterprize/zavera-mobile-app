"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeft,
  MessageSquare,
  Send,
  RefreshCcw,
  CheckCircle,
  XCircle,
  Truck,
  Search,
  FileText,
  User,
  Clock,
  AlertTriangle,
} from "lucide-react";
import {
  getDisputeById,
  getDisputeMessages,
  startDisputeInvestigation,
  requestDisputeEvidence,
  resolveDispute,
  closeDispute,
  addDisputeMessage,
  Dispute,
  DisputeMessage,
} from "@/lib/adminApi";
import { useDialog } from "@/context/DialogContext";

const statusColors: Record<string, string> = {
  OPEN: "bg-amber-500/20 text-amber-400 border-amber-500/30",
  INVESTIGATING: "bg-purple-500/20 text-purple-400 border-purple-500/30",
  EVIDENCE_REQUIRED: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  PENDING_RESOLUTION: "bg-cyan-500/20 text-cyan-400 border-cyan-500/30",
  RESOLVED_REFUND: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
  RESOLVED_RESHIP: "bg-teal-500/20 text-teal-400 border-teal-500/30",
  RESOLVED_REJECTED: "bg-red-500/20 text-red-400 border-red-500/30",
  CLOSED: "bg-gray-500/20 text-gray-400 border-gray-500/30",
};

export default function DisputeDetailPage() {
  const dialog = useDialog();
  const params = useParams();
  const router = useRouter();
  const disputeId = parseInt(params.id as string);

  const [dispute, setDispute] = useState<Dispute | null>(null);
  const [messages, setMessages] = useState<DisputeMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [newMessage, setNewMessage] = useState("");
  const [isInternal, setIsInternal] = useState(false);
  const [showResolveModal, setShowResolveModal] = useState(false);
  const [resolveData, setResolveData] = useState({
    resolution: "RESOLVED_REFUND",
    resolution_notes: "",
    create_refund: false,
    create_reship: false,
  });

  useEffect(() => {
    loadDispute();
  }, [disputeId]);

  const loadDispute = async () => {
    try {
      const [disputeData, messagesData] = await Promise.all([
        getDisputeById(disputeId),
        getDisputeMessages(disputeId, true),
      ]);
      setDispute(disputeData);
      setMessages(messagesData);
    } catch (error) {
      console.error("Failed to load dispute:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleStartInvestigation = async () => {
    setActionLoading("investigate");
    try {
      await startDisputeInvestigation(disputeId);
      await loadDispute();
    } catch (error) {
      console.error("Failed to start investigation:", error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleRequestEvidence = async () => {
    const message = prompt("Enter message to customer requesting evidence:");
    if (!message) return;

    setActionLoading("evidence");
    try {
      await requestDisputeEvidence(disputeId, message);
      await loadDispute();
    } catch (error) {
      console.error("Failed to request evidence:", error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleResolve = async () => {
    if (!resolveData.resolution_notes.trim()) {
      await dialog.alert({
        title: 'Catatan Diperlukan',
        message: 'Silakan masukkan catatan resolusi',
        variant: 'warning'
      });
      return;
    }

    setActionLoading("resolve");
    try {
      await resolveDispute(disputeId, resolveData);
      setShowResolveModal(false);
      await loadDispute();
    } catch (error) {
      console.error("Failed to resolve dispute:", error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleClose = async () => {
    const confirmed = await dialog.confirm({
      title: 'Tutup Dispute',
      message: 'Apakah Anda yakin ingin menutup dispute ini?',
      variant: 'warning',
      confirmText: 'Ya, Tutup',
      cancelText: 'Batal'
    });
    
    if (!confirmed) return;

    setActionLoading("close");
    try {
      await closeDispute(disputeId);
      await loadDispute();
    } catch (error) {
      console.error("Failed to close dispute:", error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleSendMessage = async () => {
    if (!newMessage.trim()) return;

    setActionLoading("message");
    try {
      await addDisputeMessage(disputeId, { message: newMessage, is_internal: isInternal });
      setNewMessage("");
      await loadDispute();
    } catch (error) {
      console.error("Failed to send message:", error);
    } finally {
      setActionLoading(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
      </div>
    );
  }

  if (!dispute) {
    return (
      <div className="flex flex-col items-center justify-center h-96">
        <AlertTriangle className="text-amber-400 mb-4" size={48} />
        <p className="text-white text-lg">Dispute not found</p>
      </div>
    );
  }

  const canInvestigate = dispute.status === "OPEN";
  const canRequestEvidence = ["OPEN", "INVESTIGATING"].includes(dispute.status);
  const canResolve = ["OPEN", "INVESTIGATING", "EVIDENCE_REQUIRED", "PENDING_RESOLUTION"].includes(dispute.status);
  const canClose = dispute.status.startsWith("RESOLVED_");
  const isClosed = dispute.status === "CLOSED";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => router.back()}
          className="p-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-white">{dispute.dispute_code}</h1>
            <span className={`px-3 py-1 rounded-lg text-sm font-medium border ${statusColors[dispute.status]}`}>
              {dispute.status.replace(/_/g, " ")}
            </span>
          </div>
          <p className="text-white/60 mt-1">{dispute.title}</p>
        </div>
      </div>

      {/* Actions */}
      {!isClosed && (
        <div className="flex flex-wrap gap-3">
          {canInvestigate && (
            <button
              onClick={handleStartInvestigation}
              disabled={actionLoading !== null}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-purple-500/20 text-purple-400 hover:bg-purple-500/30 transition-colors disabled:opacity-50"
            >
              <Search size={18} />
              Start Investigation
            </button>
          )}
          {canRequestEvidence && (
            <button
              onClick={handleRequestEvidence}
              disabled={actionLoading !== null}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors disabled:opacity-50"
            >
              <FileText size={18} />
              Request Evidence
            </button>
          )}
          {canResolve && (
            <button
              onClick={() => setShowResolveModal(true)}
              disabled={actionLoading !== null}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30 transition-colors disabled:opacity-50"
            >
              <CheckCircle size={18} />
              Resolve
            </button>
          )}
          {canClose && (
            <button
              onClick={handleClose}
              disabled={actionLoading !== null}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-gray-500/20 text-gray-400 hover:bg-gray-500/30 transition-colors disabled:opacity-50"
            >
              <XCircle size={18} />
              Close Dispute
            </button>
          )}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Description */}
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Description</h2>
            <p className="text-white/80 whitespace-pre-wrap">{dispute.description}</p>
            {dispute.customer_claim && (
              <div className="mt-4 p-4 rounded-xl bg-amber-500/10 border border-amber-500/20">
                <p className="text-amber-400 text-sm font-medium mb-1">Customer Claim</p>
                <p className="text-white/80">{dispute.customer_claim}</p>
              </div>
            )}
          </div>

          {/* Messages */}
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Communication</h2>

            <div className="space-y-4 max-h-96 overflow-y-auto mb-4">
              {messages.length === 0 ? (
                <p className="text-white/40 text-center py-8">No messages yet</p>
              ) : (
                messages.map((msg) => (
                  <div
                    key={msg.id}
                    className={`p-4 rounded-xl ${
                      msg.sender_type === "admin"
                        ? "bg-blue-500/10 border border-blue-500/20 ml-8"
                        : msg.sender_type === "system"
                        ? "bg-white/5 border border-white/10"
                        : "bg-white/5 border border-white/10 mr-8"
                    } ${msg.is_internal ? "border-dashed" : ""}`}
                  >
                    <div className="flex items-center gap-2 mb-2">
                      <span
                        className={`text-sm font-medium ${
                          msg.sender_type === "admin"
                            ? "text-blue-400"
                            : msg.sender_type === "system"
                            ? "text-white/40"
                            : "text-white"
                        }`}
                      >
                        {msg.sender_name || msg.sender_type}
                      </span>
                      {msg.is_internal && (
                        <span className="px-2 py-0.5 rounded text-xs bg-yellow-500/20 text-yellow-400">Internal</span>
                      )}
                      <span className="text-white/40 text-xs ml-auto">{formatDate(msg.created_at)}</span>
                    </div>
                    <p className="text-white/80">{msg.message}</p>
                  </div>
                ))
              )}
            </div>

            {/* New Message */}
            {!isClosed && (
              <div className="border-t border-white/10 pt-4">
                <div className="flex items-center gap-2 mb-3">
                  <label className="flex items-center gap-2 text-sm text-white/60 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={isInternal}
                      onChange={(e) => setIsInternal(e.target.checked)}
                      className="rounded border-white/20"
                    />
                    Internal note (not visible to customer)
                  </label>
                </div>
                <div className="flex gap-3">
                  <input
                    type="text"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                    placeholder="Type a message..."
                    className="flex-1 px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30"
                    onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
                  />
                  <button
                    onClick={handleSendMessage}
                    disabled={!newMessage.trim() || actionLoading === "message"}
                    className="px-4 py-3 rounded-xl bg-white text-black font-medium hover:bg-white/90 transition-colors disabled:opacity-50"
                  >
                    <Send size={20} />
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Customer Info */}
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Customer</h2>
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <User className="text-white/40" size={18} />
                <span className="text-white">{dispute.customer_email}</span>
              </div>
              <div className="flex items-center gap-3">
                <Clock className="text-white/40" size={18} />
                <span className="text-white/60">{formatDate(dispute.created_at)}</span>
              </div>
            </div>
          </div>

          {/* Dispute Info */}
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Details</h2>
            <div className="space-y-4">
              <div>
                <p className="text-white/40 text-sm">Type</p>
                <p className="text-white">{dispute.dispute_type.replace(/_/g, " ")}</p>
              </div>
              {dispute.order_code && (
                <div>
                  <p className="text-white/40 text-sm">Order</p>
                  <p className="text-white">{dispute.order_code}</p>
                </div>
              )}
              {dispute.shipment_id && (
                <div>
                  <p className="text-white/40 text-sm">Shipment ID</p>
                  <p className="text-white">#{dispute.shipment_id}</p>
                </div>
              )}
              {dispute.resolution && (
                <div>
                  <p className="text-white/40 text-sm">Resolution</p>
                  <p className="text-white">{dispute.resolution.replace(/_/g, " ")}</p>
                </div>
              )}
              {dispute.resolution_notes && (
                <div>
                  <p className="text-white/40 text-sm">Resolution Notes</p>
                  <p className="text-white/80 text-sm">{dispute.resolution_notes}</p>
                </div>
              )}
            </div>
          </div>

          {/* Evidence */}
          {dispute.evidence_urls && dispute.evidence_urls.length > 0 && (
            <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
              <h2 className="text-lg font-semibold text-white mb-4">Evidence</h2>
              <div className="space-y-2">
                {dispute.evidence_urls.map((url, index) => (
                  <a
                    key={index}
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block p-3 rounded-xl bg-white/5 text-white/60 hover:text-white hover:bg-white/10 transition-colors truncate"
                  >
                    Evidence {index + 1}
                  </a>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Resolve Modal */}
      {showResolveModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6 w-full max-w-md">
            <h3 className="text-xl font-bold text-white mb-4">Resolve Dispute</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-white/60 text-sm mb-2">Resolution Type</label>
                <select
                  value={resolveData.resolution}
                  onChange={(e) => setResolveData({ ...resolveData, resolution: e.target.value })}
                  className="w-full px-4 py-3 rounded-xl bg-neutral-900 border border-white/10 text-white focus:outline-none focus:border-white/30 hover:border-white/20 transition-colors cursor-pointer"
                >
                  <option value="RESOLVED_REFUND">Refund</option>
                  <option value="RESOLVED_RESHIP">Reship</option>
                  <option value="RESOLVED_REJECTED">Reject</option>
                </select>
              </div>

              <div>
                <label className="block text-white/60 text-sm mb-2">Resolution Notes</label>
                <textarea
                  value={resolveData.resolution_notes}
                  onChange={(e) => setResolveData({ ...resolveData, resolution_notes: e.target.value })}
                  placeholder="Enter resolution details..."
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30 resize-none h-32"
                />
              </div>

              {resolveData.resolution === "RESOLVED_REFUND" && (
                <label className="flex items-center gap-2 text-white/60 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={resolveData.create_refund}
                    onChange={(e) => setResolveData({ ...resolveData, create_refund: e.target.checked })}
                    className="rounded border-white/20"
                  />
                  Auto-create refund
                </label>
              )}

              {resolveData.resolution === "RESOLVED_RESHIP" && (
                <label className="flex items-center gap-2 text-white/60 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={resolveData.create_reship}
                    onChange={(e) => setResolveData({ ...resolveData, create_reship: e.target.checked })}
                    className="rounded border-white/20"
                  />
                  Auto-create reship
                </label>
              )}
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowResolveModal(false)}
                className="flex-1 px-4 py-3 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleResolve}
                disabled={!resolveData.resolution_notes.trim() || actionLoading === "resolve"}
                className="flex-1 px-4 py-3 rounded-xl bg-emerald-500 text-white font-medium hover:bg-emerald-600 transition-colors disabled:opacity-50"
              >
                {actionLoading === "resolve" ? "Processing..." : "Resolve"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
