import logoUrl from "../../../ui/icon.svg";

export function Logo() {
  return <div className="flex items-center gap-3">
    <div className="logo-mark"><img src={logoUrl} alt="" /></div>
    <div><div className="text-[15px] font-semibold tracking-[-0.02em] text-white">MK-Link</div>
      <div className="text-[10px] font-medium uppercase tracking-[0.18em] text-slate-500">Windows utility</div></div>
  </div>;
}
