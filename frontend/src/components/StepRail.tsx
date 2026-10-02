import { Check, Circle, MapPin, Pencil, ShieldCheck, Target, Waypoints } from "lucide-react";

const steps = [["Target", Target], ["Link type", Waypoints], ["Name", Pencil], ["Destination", MapPin], ["Review", ShieldCheck]] as const;

export function StepRail({ current }: { current: number }) {
  return <aside className="w-[210px] shrink-0 border-r border-white/[0.06] bg-[#0b0f15]/90 px-4 py-6">
    <div className="mb-8 px-2"><div className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-600">Create a link</div>
      <div className="mt-1 text-xs text-slate-500">A safe, guided workflow</div></div>
    <div className="space-y-1">{steps.map(([label, Icon], index) => {
      const active=index===current, complete=index<current;
      return <div key={label} className="relative">{index<steps.length-1 && <div className="absolute left-[18px] top-[39px] h-[27px] w-px bg-white/[0.06]" />}
        <div className={"step-item "+(active?"step-active":"")}><div className={"step-icon "+(complete?"step-complete":active?"step-current":"")}>
          {complete?<Check size={14}/>:active?<Icon size={15}/>:<Circle size={10}/>}</div>
          <div><div className={"text-[12px] font-medium "+(active?"text-slate-100":"text-slate-500")}>{label}</div>
            <div className="mt-0.5 text-[10px] text-slate-700">{complete?"Complete":active?"Current step":"Upcoming"}</div></div>
        </div>
      </div>;
    })}</div>
  </aside>;
}
