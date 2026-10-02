import { Check, FolderGit2, GitBranch, Link2 } from "lucide-react";
import type { LinkTypeInfo } from "../types";

const icons={symbolic:Link2,hard:GitBranch,junction:FolderGit2} as const;

export function TypeCard({item,selected,onClick}:{item:LinkTypeInfo;selected:boolean;onClick:()=>void}) {
  const Icon=icons[item.ID];
  return <button onClick={onClick} className={"type-card text-left "+(selected?"type-card-selected":"")}>
    <div className="flex items-start gap-4"><div className="type-icon"><Icon size={19}/></div><div className="min-w-0 flex-1">
      <div className="flex items-center gap-2"><div className="font-semibold text-slate-100">{item.Name}</div>
        {selected&&<span className="badge"><Check size={11}/> Selected</span>}</div>
      <div className="mt-1 text-[12px] leading-5 text-slate-500">{item.Summary}</div></div></div>
    <div className="mt-4 grid grid-cols-2 gap-3 border-t border-white/[0.05] pt-3">
      <div><div className="meta-label">Best for</div><div className="mt-1 text-[11px] leading-4 text-slate-400">{item.BestFor}</div></div>
      <div><div className="meta-label">Works with</div><div className="mt-1 text-[11px] leading-4 text-slate-400">{item.TargetKind}</div></div>
    </div>
  </button>;
}
