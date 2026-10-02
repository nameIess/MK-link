import { ArrowDown, CheckCircle2, File, Folder, Link2 } from "lucide-react";
import type { LinkTypeInfo } from "../types";

export function Preview({target,destination,name,typeInfo}:{target:string;destination:string;name:string;typeInfo?:LinkTypeInfo}) {
  const isDir=typeInfo?.TargetKind.includes("directories")||typeInfo?.TargetKind.includes("folders");
  const link=destination&&name?destination.replace(/[\\/]$/,"")+"\\"+name:"Choose a destination";
  return <aside className="w-[310px] shrink-0 border-l border-white/[0.06] bg-[#0a0e14] px-5 py-6">
    <div className="mb-5"><div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-600">Live preview</div>
      <div className="mt-1 text-sm font-medium text-slate-300">What MK-Link will create</div></div>
    <div className="space-y-2"><div className="preview-node"><div className="icon-tile small">{isDir?<Folder size={15}/>:<File size={15}/>}</div>
      <div className="min-w-0"><div className="meta-label">Target</div><div className="mt-1 truncate font-mono text-[10px] text-slate-400">{target||"Not selected"}</div></div></div>
      <div className="flex justify-center py-1 text-slate-700"><ArrowDown size={15}/></div>
      <div className="preview-node preview-link"><div className="icon-tile small accent"><Link2 size={15}/></div>
        <div className="min-w-0"><div className="meta-label text-blue-400/70">New link</div><div className="mt-1 truncate font-mono text-[10px] text-slate-300">{link}</div></div></div></div>
    <div className="mt-6 rounded-xl border border-white/[0.05] bg-white/[0.015] p-4"><div className="meta-label">Link type</div>
      <div className="mt-2 text-sm font-medium text-slate-200">{typeInfo?.Name||"Not selected"}</div>
      <div className="mt-2 text-[11px] leading-5 text-slate-500">{typeInfo?.Description||"Choose a link type to see how it behaves."}</div></div>
    <div className="mt-4 flex items-center gap-2 text-[10px] text-slate-600"><CheckCircle2 size={13}/> Nothing is copied or moved</div>
  </aside>;
}
