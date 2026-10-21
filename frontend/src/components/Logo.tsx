export function Logo() {
  return <div className="flex items-center gap-3">
    <div className="logo-mark">
      <svg viewBox="0 0 128 128" role="img" aria-label="MK-Link" xmlns="http://www.w3.org/2000/svg">
        <g fill="none" stroke="currentColor" strokeWidth="12" strokeLinecap="round" strokeLinejoin="round">
          <path d="M51 42 41 32a19 19 0 0 0-27 27l19 19a19 19 0 0 0 27 0l10-10" />
          <path d="m77 86 10 10a19 19 0 0 0 27-27L95 50a19 19 0 0 0-27 0L58 60" />
        </g>
      </svg>
    </div>
    <div><div className="text-[15px] font-semibold tracking-[-0.02em] text-white">MK-Link</div>
      <div className="text-[10px] font-medium uppercase tracking-[0.18em] text-slate-500">Windows utility</div></div>
  </div>;
}
