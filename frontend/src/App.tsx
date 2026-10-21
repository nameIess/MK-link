import { useEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { ArrowLeft, ArrowRight, Check, ChevronRight, FolderOpen, Info, ShieldAlert, Sparkles, X } from "lucide-react";
import { Logo } from "./components/Logo";
import { StepRail } from "./components/StepRail";
import { PathCard } from "./components/PathCard";
import { Preview } from "./components/Preview";
import { TypeCard } from "./components/TypeCard";
import { chooseDestination, chooseTargetDirectory, chooseTargetFile, createLink, getTypes, inspectTarget, openInExplorer, validateDestination } from "./api";
import type { CreateLinkResult, LinkType, LinkTypeInfo } from "./types";

declare global {
  interface Window {
    runtime?: {
      OnFileDrop?: (callback: (x: number, y: number, paths: string[]) => void, useDropTarget: boolean) => void;
      OnFileDropOff?: () => void;
    };
  }
}

export default function App() {
  const [step,setStep]=useState(0);
  const [target,setTarget]=useState("");
  const [targetKind,setTargetKind]=useState<"file"|"folder">("folder");
  const [types,setTypes]=useState<LinkTypeInfo[]>([]);
  const [type,setType]=useState<LinkType|"">("");
  const [name,setName]=useState("");
  const [destination,setDestination]=useState("");
  const [error,setError]=useState("");
  const [busy,setBusy]=useState(false);
  const [result,setResult]=useState<CreateLinkResult|null>(null);
  const [showAbout,setShowAbout]=useState(false);
  const flowRequestId = useRef(0);

  const selectedType=useMemo(()=>types.find(x=>x.ID===type),[types,type]);
  const linkPath=destination&&name?destination.replace(/[\\/]$/,"")+"\\"+name.trim():"";
  const canNext=step===0?!!target:step===1?!!type:step===2?!!name.trim():step===3?!!destination:true;

  useEffect(()=>{
    const drop=async (_x:number,_y:number,paths:string[])=>{ if(paths[0]) await applyTarget(paths[0]); };
    window.runtime?.OnFileDrop?.(drop,true);
    return ()=>window.runtime?.OnFileDropOff?.();
  },[]);

  useEffect(()=>{
    if(!showAbout)return;
    const onKeyDown=(event:KeyboardEvent)=>{if(event.key==="Escape")setShowAbout(false);};
    window.addEventListener("keydown",onKeyDown);
    return ()=>window.removeEventListener("keydown",onKeyDown);
  },[showAbout]);

  function invalidatePending(){ flowRequestId.current++; }

  async function applyTarget(path:string){
    const requestId=++flowRequestId.current;
    setError("");
    try{
      const info=await inspectTarget(path);
      if(requestId!==flowRequestId.current)return;
      const next=await getTypes(info.Path);
      if(requestId!==flowRequestId.current)return;
      setTarget(info.Path);
      setTargetKind(info.IsDir?"folder":"file");
      setTypes(next);
      setType(next[0]?.ID||"");
      const base=info.Path.split(/[\\/]/).filter(Boolean).pop()||"";
      setName(base);
      setDestination("");
      setResult(null);
      setStep(1);
    }catch(e){
      if(requestId===flowRequestId.current)setError(String(e));
    }
  }

  async function chooseFile(){
    const requestId=++flowRequestId.current;
    setError("");
    try{
      const path=await chooseTargetFile();
      if(requestId!==flowRequestId.current||!path)return;
      await applyTarget(path);
    }catch(e){if(requestId===flowRequestId.current)setError(String(e));}
  }

  async function chooseFolder(){
    const requestId=++flowRequestId.current;
    setError("");
    try{
      const path=await chooseTargetDirectory();
      if(requestId!==flowRequestId.current||!path)return;
      await applyTarget(path);
    }catch(e){if(requestId===flowRequestId.current)setError(String(e));}
  }

  async function chooseDest(){
    const requestId=++flowRequestId.current;
    setError("");
    try{
      const path=await chooseDestination();
      if(requestId!==flowRequestId.current||!path)return;
      setDestination(path);
    }catch(e){if(requestId===flowRequestId.current)setError(String(e));}
  }

  async function next(){
    const requestId=++flowRequestId.current;
    setError("");
    if(step===3){
      if(!type||!destination||!name.trim())return;
      try{
        const validation=await validateDestination(destination,name.trim(),type);
        if(requestId!==flowRequestId.current)return;
        if(!validation.Valid){setError(validation.Message);return;}
      }catch(e){
        if(requestId===flowRequestId.current)setError(String(e));
        return;
      }
    }
    if(requestId===flowRequestId.current)setStep(current=>Math.min(4,current+1));
  }

  async function submit(){
    if(!type||busy)return;
    const requestId=++flowRequestId.current;
    setBusy(true);
    setError("");
    try{
      const response=await createLink({TargetPath:target,DestinationDir:destination,LinkName:name.trim(),LinkType:type});
      if(requestId!==flowRequestId.current)return;
      setResult(response);
      if(!response.Success)setError(response.Error||"MK-Link could not create the link.");
    }catch(e){
      if(requestId===flowRequestId.current)setError(String(e));
    }finally{
      if(requestId===flowRequestId.current)setBusy(false);
    }
  }

  function reset(){
    invalidatePending();
    setBusy(false);
    setStep(0);setTarget("");setTypes([]);setType("");setName("");setDestination("");setError("");setResult(null);
  }

  function goBack(){
    invalidatePending();
    setError("");
    setStep(current=>Math.max(0,current-1));
  }

  function updateName(value:string){invalidatePending();setName(value);setError("");}
  function updateType(value:LinkType){invalidatePending();setType(value);setError("");}
  function updateDestination(value:string){invalidatePending();setDestination(value);setError("");}

  return <div className="app-shell">
    <header className="titlebar">
      <Logo/>
      <div className="flex items-center gap-2"><div className="status-dot"><span/>Windows</div><button className="title-action" title="About MK-Link" aria-label="About MK-Link" onClick={()=>setShowAbout(true)}><Info size={15}/></button></div>
    </header>

    <div className="flex min-h-0 flex-1">
      <StepRail current={step}/>
      <main className="min-w-0 flex-1 overflow-auto">
        <div className="mx-auto max-w-[760px] px-10 py-10">
          {step===0&&<section className="animate-in">
            <Eyebrow>01 / Target</Eyebrow><h1>Select what you want to link.</h1>
            <p className="lead">Choose a file or folder. MK-Link leaves the original untouched and creates a filesystem link at the destination you choose.</p>
            <div className="drop-zone" style={{"--wails-drop-target":"drop"} as CSSProperties}>
              <div className="drop-icon"><FolderOpen size={23}/></div>
              <div className="text-sm font-medium text-slate-200">Drag a file or folder here</div>
              <div className="mt-1 text-xs text-slate-600">or use a native Windows picker</div>
              <div className="mt-5 flex justify-center gap-2">
                <button className="btn-secondary" onClick={chooseFile}><FolderOpen size={15}/>Choose file</button>
                <button className="btn-secondary" onClick={chooseFolder}><FolderOpen size={15}/>Choose folder</button>
              </div>
            </div>
          </section>}

          {step===1&&<section className="animate-in">
            <Eyebrow>02 / Link type</Eyebrow><h1>Choose how the link should behave.</h1>
            <p className="lead">Each type solves a different filesystem problem. The cards explain what each is best for before you choose.</p>
            <div className="space-y-3">{types.map(item=><TypeCard key={item.ID} item={item} selected={type===item.ID} onClick={()=>updateType(item.ID)}/>)}</div>
          </section>}

          {step===2&&<section className="animate-in">
            <Eyebrow>03 / Name</Eyebrow><h1>Give the link a clear name.</h1>
            <p className="lead">This is the filename or folder name that appears at the destination. It does not rename or move the target.</p>
            <PathCard path={target} kind={targetKind}/>
            <label className="field-label" htmlFor="link-name">Link name</label>
            <input id="link-name" autoFocus value={name} onChange={e=>updateName(e.target.value)} onKeyDown={e=>e.key==="Enter"&&next()} className="text-input" placeholder="MyProject" maxLength={240}/>
            <div className="mt-3 flex gap-2 text-[11px] text-slate-600"><Info size={14}/>Use a normal Windows filename. Path separators and alternate data stream syntax are blocked.</div>
          </section>}

          {step===3&&<section className="animate-in">
            <Eyebrow>04 / Destination</Eyebrow><h1>Where should the link live?</h1>
            <p className="lead">Choose the folder where <span className="text-slate-300">{name||"your link"}</span> will be created.</p>
            <button className="destination-picker" onClick={chooseDest}><FolderOpen size={18}/>
              <div className="min-w-0 flex-1 text-left"><div className="text-xs font-medium text-slate-200">{destination?"Destination folder":"Choose destination folder"}</div>
                <div className="mt-1 truncate font-mono text-[11px] text-slate-600">{destination||"Use the native Windows folder picker"}</div></div><ChevronRight size={16} className="text-slate-700"/></button>
            {destination&&<div className="mt-4 rounded-xl border border-white/[0.05] bg-white/[0.015] p-4"><div className="meta-label">Resulting path</div><div className="mt-2 break-all font-mono text-xs text-slate-300">{linkPath}</div></div>}
          </section>}

          {step===4&&<section className="animate-in">
            <Eyebrow>05 / Review</Eyebrow><h1>Review before creating.</h1>
            <p className="lead">Nothing is changed until you press Create link. Existing files and folders are never overwritten automatically.</p>
            <div className="review-card"><ReviewRow label="Target" value={target}/><ReviewRow label="Link type" value={selectedType?.Name||type}/><ReviewRow label="Link name" value={name}/><ReviewRow label="Destination" value={destination}/><div className="review-divider"/><ReviewRow label="Will create" value={linkPath} accent/></div>
            <div className="notice"><ShieldAlert size={16}/><div><div className="text-xs font-medium text-slate-300">Safe by default</div><div className="mt-1 text-[11px] leading-5 text-slate-600">MK-Link uses native Windows filesystem APIs. User paths are never assembled into a shell command.</div></div></div>
          </section>}

          {result?.Success&&<section className="success-card mt-8"><div className="success-icon"><Check size={22}/></div>
            <div className="min-w-0 flex-1"><div className="text-sm font-semibold text-slate-100">Link created successfully</div><div className="mt-1 truncate font-mono text-[11px] text-slate-500">{result.LinkPath}</div></div>
            <button className="btn-secondary" onClick={()=>openInExplorer(result.LinkPath)}><FolderOpen size={14}/>Open</button></section>}

          {error&&<div className="error-card mt-5" role="alert"><ShieldAlert size={17}/><div className="min-w-0 flex-1"><div className="text-xs font-semibold text-red-300">Could not continue</div>
            <div className="mt-1 text-[11px] leading-5 text-red-300/60">{error}</div>
            {result?.NeedsElevation&&<div className="mt-3 text-[11px] text-amber-300/70">Windows requires elevation for this operation. Restart MK-Link as administrator and retry.</div>}
          </div><button className="icon-button" aria-label="Dismiss error" onClick={()=>setError("")}><X size={14}/></button></div>}
        </div>
      </main>
      <Preview target={target} destination={destination} name={name} typeInfo={selectedType}/>
    </div>

    <footer className="footerbar"><div className="text-[10px] font-medium uppercase tracking-[0.14em] text-slate-700">MK-Link · Windows filesystem utility</div>
      <div className="flex items-center gap-2">
        {step>0&&!result?.Success&&<button className="btn-ghost" disabled={busy} onClick={goBack}><ArrowLeft size={14}/>Back</button>}
        {step<4&&<button className="btn-primary" disabled={!canNext||busy} onClick={next}>Continue<ArrowRight size={15}/></button>}
        {step===4&&!result?.Success&&<button className="btn-primary" disabled={busy} onClick={submit}>{busy?<span className="spinner" aria-hidden="true"/>:<Sparkles size={15}/>} {busy?"Creating…":"Create link"}</button>}
        {result?.Success&&<button className="btn-primary" onClick={reset}>Create another<ArrowRight size={15}/></button>}
      </div>
    </footer>

    {showAbout&&<div className="dialog-backdrop" onMouseDown={event=>{if(event.target===event.currentTarget)setShowAbout(false);}}>
      <section className="about-dialog" role="dialog" aria-modal="true" aria-labelledby="about-title" aria-describedby="about-description" onKeyDown={event=>{if(event.key==="Escape"){setShowAbout(false);return;}if(event.key==="Tab"){event.preventDefault();event.currentTarget.querySelector("button")?.focus();}}}>
        <div className="flex items-start gap-3"><div className="icon-tile"><Info size={19}/></div><div className="min-w-0 flex-1"><h2 id="about-title" className="text-base font-semibold text-slate-100">About MK-Link</h2><p id="about-description" className="mt-2 text-sm leading-6 text-slate-400">A guided Windows utility for creating symbolic links, hard links, and junctions. MK-Link validates the requested operation and uses native Windows filesystem APIs.</p></div></div>
        <div className="mt-5 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3 text-xs text-slate-500">Designed to make filesystem links easier to understand and safer to create.</div>
        <div className="mt-5 flex justify-end"><button autoFocus className="btn-primary" onClick={()=>setShowAbout(false)}>Close</button></div>
      </section>
    </div>}
  </div>;
}

function Eyebrow({children}:{children:React.ReactNode}){return <div className="mb-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-blue-400/70">{children}</div>}
function ReviewRow({label,value,accent}:{label:string;value:string;accent?:boolean}){return <div className="grid grid-cols-[105px_1fr] gap-5 py-3"><div className="meta-label">{label}</div><div className={"min-w-0 break-all font-mono text-[11px] "+(accent?"text-blue-300":"text-slate-300")}>{value}</div></div>}
