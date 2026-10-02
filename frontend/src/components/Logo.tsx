import { Link2 } from "lucide-react";

export function Logo() {
  return <div className="flex items-center gap-3">
    <div className="logo-mark"><Link2 size={19} strokeWidth={2.4} /></div>
    <div><div className="text-[15px] font-semibold tracking-[-0.02em] text-white">MK-Link</div>
      <div className="text-[10px] font-medium uppercase tracking-[0.18em] text-slate-500">Windows utility</div></div>
  </div>;
}
