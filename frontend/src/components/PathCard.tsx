import { File, Folder, X } from "lucide-react";

export function PathCard({path,kind,onClear}:{path:string;kind:"file"|"folder";onClear?:()=>void}) {
  const Icon=kind==="folder"?Folder:File;
  return <div className="surface flex items-center gap-3 p-4"><div className="icon-tile"><Icon size={18}/></div>
    <div className="min-w-0 flex-1"><div className="text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-600">{kind}</div>
      <div className="mt-1 truncate font-mono text-[12px] text-slate-300" title={path}>{path}</div></div>
    {onClear&&<button className="icon-button" onClick={onClear} aria-label="Clear"><X size={15}/></button>}</div>;
}
