var W=Object.defineProperty;var Y=(t,o,n)=>o in t?W(t,o,{enumerable:!0,configurable:!0,writable:!0,value:n}):t[o]=n;var w=(t,o,n)=>(Y(t,typeof o!="symbol"?o+"":o,n),n);import{d as Z,aa as X,C as y,J as T,a4 as I,o as v,c as m,m as f,f as d,e as q,p as k,K as A,am as ee,r as te,t as S,an as ne,ao as se,n as B,F as ie,G as oe,q as N,ap as le,aq as ae,ar as re,as as ue,v as ce,x as de,_ as fe}from"./index-CvRMgvyl.js";function ge(t,o,n){return Math.max(o,Math.min(t,n))}const pe=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class ve{constructor(o,n){w(this,"commands");w(this,"keyMap");w(this,"boundTriggerShortcuts");this.commands=n,this.keyMap=Object.fromEntries(Object.entries(o).map(([h,r])=>[h.toLowerCase(),r])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(o){me(o,this.keyMap,this.commands)}}function me(t,o,n){const h=he(t.code),r=[t.ctrlKey?"ctrl":"",t.shiftKey?"shift":"",t.altKey?"alt":"",h].filter(b=>b!=="").join("+"),i=o[r];if(!i)return;const s=n[i];s.isAllowedContext&&!s.isAllowedContext(t)||(s.shouldPreventDefaultAction&&t.preventDefault(),!(s.isDisabled&&s.isDisabled())&&s.trigger(t))}function he(t){return pe.includes(t)?"":t.replace(/^Key/,"").toLowerCase()}function be(t,o){const n=" "+t,h=n.matchAll(/ ([-\s\w]+):\s*/g),r=[];for(const i of Array.from(h)){if(i.index===void 0)continue;const s=ye(i[1]);if(o.length>0&&!o.includes(s))throw new Error(`Unknown field “${s}”. Known fields: ${o.join(", ")}`);const b=i.index+i[0].length,u=n.substring(b);let c;if(/^\s*["']/.test(u)){const g=u.match(/['"](.*?)['"]/);if(g!==null)c=g[1];else throw new Error(`Quote mismatch for field “${s}”.`)}else{const g=u.indexOf(" "),_=g===-1?u.length:g;c=u.substring(0,_)}c!==""&&r.push([s,c])}return r}function ye(t){return t.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(o,n)=>n===0?o:o.substring(1).toUpperCase())}const j=t=>(ce("data-v-e7ca9a15"),t=t(),de(),t),ke=j(()=>f("span",{class:"visually-hidden"},"Focus filter",-1)),Se={class:"k-filter-icon"},_e=["for"],Ce=["id","placeholder"],xe={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},we={class:"k-suggestion-list"},Te={key:0,class:"k-filter-bar-error"},Fe={key:0},Ie=["title","data-filter-field"],qe={class:"visually-hidden"},Ae=j(()=>f("span",{class:"visually-hidden"},"Clear query",-1)),Ne=Z({__name:"FilterBar",props:{id:{type:String,required:!1,default:()=>X("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(t,{emit:o}){const n=t,h=o,r=y(null),i=y(null),s=y(n.query),b=y([]),u=y(null),c=y(!1),p=y(-1),g=T(()=>Object.keys(n.fields)),_=T(()=>Object.entries(n.fields).slice(0,5).map(([e,l])=>({fieldName:e,...l}))),E=T(()=>g.value.length>0?`Filter by ${g.value.join(", ")}`:"Filter"),O=T(()=>n.placeholder??E.value);I(()=>b.value,function(e,l){G(e,l)||(u.value=null,h("fields-change",{fields:e,query:s.value}))}),I(()=>n.query,()=>{s.value=n.query,C(s.value)},{immediate:!0}),I(()=>s.value,function(){s.value===""&&(u.value=null)});const D={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},$={submitQuery:{trigger:L,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:K,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:V,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:F,isAllowedContext(e){return r.value!==null&&e.composedPath().includes(r.value)}}};function z(){const e=new ve(D,$);re(function(){e.registerListener()}),ue(function(){e.unRegisterListener()}),C(s.value)}z();function Q(e){const l=e.target;C(l.value)}function L(){if(i.value instanceof HTMLInputElement)if(p.value===-1)C(i.value.value),c.value=!1;else{const e=_.value[p.value].fieldName;e&&P(i.value,e)}}function K(){M(1)}function V(){M(-1)}function M(e){p.value=ge(p.value+e,-1,_.value.length-1)}function U(){i.value instanceof HTMLInputElement&&i.value.focus()}function H(e){const a=e.currentTarget.getAttribute("data-filter-field");a&&i.value instanceof HTMLInputElement&&P(i.value,a)}function P(e,l){const a=s.value===""||s.value.endsWith(" ")?"":" ";s.value+=a+l+":",e.focus(),p.value=-1}function R(){s.value="",i.value instanceof HTMLInputElement&&(i.value.value="",i.value.focus(),C(""))}function J(e){e.relatedTarget===null&&F(),r.value instanceof HTMLElement&&e.relatedTarget instanceof Node&&!r.value.contains(e.relatedTarget)&&F()}function F(){c.value=!1}function C(e){u.value=null;try{const l=be(e,g.value);l.sort((a,x)=>a[0].localeCompare(x[0])),b.value=l}catch(l){if(l instanceof Error)u.value=l,c.value=!0;else throw l}}function G(e,l){return JSON.stringify(e)===JSON.stringify(l)}return(e,l)=>(v(),m("div",{ref_key:"filterBar",ref:r,class:"k-filter-bar","data-testid":"k-filter-bar"},[f("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:U},[ke,d(),f("span",Se,[q(k(ee),{decorative:"","data-testid":"k-filter-bar-filter-icon",size:k(A)},null,8,["size"])])]),d(),f("label",{for:`${n.id}-filter-bar-input`,class:"visually-hidden"},[te(e.$slots,"default",{},()=>[d(S(E.value),1)],!0)],8,_e),d(),ne(f("input",{id:`${n.id}-filter-bar-input`,ref_key:"filterInput",ref:i,"onUpdate:modelValue":l[0]||(l[0]=a=>s.value=a),class:"k-filter-bar-input",type:"text",placeholder:O.value,"data-testid":"k-filter-bar-filter-input",onFocus:l[1]||(l[1]=a=>c.value=!0),onBlur:J,onChange:Q},null,40,Ce),[[se,s.value]]),d(),c.value?(v(),m("div",xe,[f("div",we,[u.value!==null?(v(),m("p",Te,S(u.value.message),1)):(v(),m("button",{key:1,class:B(["k-submit-query-button",{"k-submit-query-button-is-selected":p.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:L},`
          Submit `+S(s.value),3)),d(),(v(!0),m(ie,null,oe(_.value,(a,x)=>(v(),m("div",{key:`${n.id}-${x}`,class:B(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":p.value===x}])},[f("b",null,S(a.fieldName),1),a.description!==""?(v(),m("span",Fe,": "+S(a.description),1)):N("",!0),d(),f("button",{class:"k-apply-suggestion-button",title:`Add ${a.fieldName}:`,type:"button","data-filter-field":a.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:H},[f("span",qe,"Add "+S(a.fieldName)+":",1),d(),q(k(le),{decorative:"",size:k(A)},null,8,["size"])],8,Ie)],2))),128))])])):N("",!0),d(),s.value!==""?(v(),m("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:R},[Ae,d(),q(k(ae),{decorative:"",size:k(A)},null,8,["size"])])):N("",!0)],512))}}),Me=fe(Ne,[["__scopeId","data-v-e7ca9a15"]]);export{Me as F};
